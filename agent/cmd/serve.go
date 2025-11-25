//go:generate cd ui && bun run build
package cmd

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"sync"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/sse"
	"github.com/google/uuid"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/agent/config"
	"github.com/situation-sh/situation/pkg/store"
	"github.com/urfave/cli/v3"
)

//go:embed all:ui/build/client
var dist embed.FS

var serveCmd = cli.Command{
	Name:   "serve",
	Usage:  "Start UI",
	Action: serveAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "addr",
			Usage: "Address to listen on",
			Value: "127.0.0.1",
		},
		&cli.IntFlag{
			Name:  "port",
			Usage: "Port to listen on",
			Value: 8888,
		},
	},
}

type AppState struct {
	mu             sync.Mutex
	InternetAccess bool   `json:"internet_access" example:"true" doc:"Whether the application has internet access"`
	Status         string `json:"status" example:"running" doc:"Application status"`
}

var appState = AppState{
	Status: "idle",
}

func (s *AppState) GetStatus() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Status
}

func (s *AppState) SetStatus(status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = status
}

func (s *AppState) UpdateInternetAccess() {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Simple check: try to resolve a well-known domain
	_, err := net.LookupHost("situation.sh")
	s.InternetAccess = err == nil
}

func (s *AppState) Copy() AppState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return AppState{
		InternetAccess: s.InternetAccess,
		Status:         s.Status,
	}
}

type StatusOutput struct {
	Body AppState
}

type SSELogDispatcher struct {
	formatter logrus.Formatter
	channels  map[string]chan []byte
	mu        sync.Mutex
}

func (d *SSELogDispatcher) Fire(entry *logrus.Entry) error {
	if entry == nil {
		return nil
	}
	bytes, err := d.formatter.Format(entry)
	if err != nil {
		return err
	}
	fmt.Println(string(bytes))
	// send log entry to sse clients
	for _, channel := range d.channels {
		channel <- bytes
	}
	return nil
}

func (d *SSELogDispatcher) Levels() []logrus.Level {
	return logrus.AllLevels
}

var serverLogger = logrus.New()

func LogrusMiddleware(ctx huma.Context, next func(huma.Context)) {
	// Set a custom header on the response.
	serverLogger.WithField("method", ctx.Method()).
		WithField("path", ctx.URL().Path).
		WithField("remote", ctx.RemoteAddr()).Info()

	// Call the next middleware in the chain. This eventually calls the
	// operation handler as well.
	next(ctx)
	if status := ctx.Status(); status >= 400 {
		serverLogger.WithField("status", status).Error()
	}
}

func initAPI() (huma.API, *http.ServeMux, *logrus.Logger) {
	// CORS: adapte l'origine à TON front (localhost:3000 si Next.js dev)

	router := http.NewServeMux()
	config := huma.DefaultConfig("Situation Agent API", Version)
	api := humago.New(router, config)

	logger := logrus.New()
	bindAPI(api, logger)
	return api, router, logger
}

func serveAction(ctx context.Context, cmd *cli.Command) error {
	api, router, _ := initAPI()
	subfs, err := fs.Sub(dist, "dist/ui")
	if err != nil {
		return err
	}
	router.Handle("/ui", http.FileServerFS(subfs))
	api.UseMiddleware(LogrusMiddleware)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"}, // ou http://127.0.0.1:3000
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodDelete},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true, // mets à false si tu n’en as pas besoin
	})

	laddr := net.JoinHostPort(cmd.String("addr"), fmt.Sprintf("%v", cmd.Int("port")))
	fmt.Printf("Starting agent on http://%s\n", laddr)
	return http.ListenAndServe(
		laddr,
		c.Handler(router),
	)
}

type NetworkInterface struct {
	Name    string   `json:"name" example:"eth0"`
	MAC     string   `json:"mac" example:"00:1A:2B:3C:4D:5E"`
	Addrs   []string `json:"addrs"`
	Running bool     `json:"running" example:"true"`
}

type NetworkOutput struct {
	Body struct {
		Interfaces []*NetworkInterface `json:"interfaces" doc:"List of up network interfaces"`
	}
}

type RawJSONOutput struct {
	Body json.RawMessage
}

type SSEMessage json.RawMessage

func bindAPI(api huma.API, logger *logrus.Logger) {
	dispatcher := SSELogDispatcher{
		channels:  make(map[string]chan []byte),
		formatter: &logrus.JSONFormatter{PrettyPrint: true},
	}
	logger.AddHook(&dispatcher)

	huma.Get(api, "/status", func(ctx context.Context, input *struct{}) (*StatusOutput, error) {
		out := &StatusOutput{}
		appState.UpdateInternetAccess()
		out.Body = appState.Copy()
		return out, nil
	})

	huma.Get(api, "/network", func(ctx context.Context, input *struct{}) (*NetworkOutput, error) {
		interfaces, err := net.Interfaces()
		if err != nil {
			return nil, err
		}
		out := &NetworkOutput{}
		out.Body.Interfaces = make([]*NetworkInterface, 0)

		for _, iface := range interfaces {
			if iface.Flags&net.FlagLoopback > 0 {
				continue
			}
			if iface.Flags&net.FlagUp == 0 {
				continue
			}

			ni := NetworkInterface{
				Name:    iface.Name,
				MAC:     iface.HardwareAddr.String(),
				Addrs:   []string{},
				Running: (iface.Flags & net.FlagRunning) > 0,
			}
			if addrs, err := iface.Addrs(); err == nil {
				for _, addr := range addrs {
					ni.Addrs = append(ni.Addrs, addr.String())
				}
			}
			out.Body.Interfaces = append(out.Body.Interfaces, &ni)
		}

		return out, nil
	})

	huma.Post(api, "/run", func(ctx context.Context, input *struct{}) (*struct{}, error) {
		s := store.NewMemoryStore(ID)
		appState.SetStatus("running")
		defer appState.SetStatus("idle")
		return nil, run(s, logger)
	})

	sse.Register(api, huma.Operation{
		OperationID: "sse-logs",
		Method:      http.MethodGet,
		Path:        "/sse/logs",
		Summary:     "Agent live logs",
	}, map[string]any{
		// Mapping of event type name to Go struct for that event.
		"message": json.RawMessage{},
	}, func(ctx context.Context, input *struct{}, send sse.Sender) {
		clientID := uuid.New().String()
		channel := make(chan []byte)

		dispatcher.mu.Lock()
		dispatcher.channels[clientID] = channel
		dispatcher.mu.Unlock()

		// Cleanup when function exits (either on disconnect or error)
		defer func() {
			dispatcher.mu.Lock()
			delete(dispatcher.channels, clientID)
			dispatcher.mu.Unlock()
			close(channel)
		}()

		for {
			select {
			case <-ctx.Done():
				// Client disconnected, exit gracefully
				return
			case raw := <-channel:
				if err := send.Data(json.RawMessage(raw)); err != nil {
					return
				}
			}
		}
	})

	huma.Get(api, "/configuration", func(ctx context.Context, input *struct{}) (*RawJSONOutput, error) {
		bytes, err := config.JSON()
		if err != nil {
			return nil, err
		}
		s := RawJSONOutput{json.RawMessage(bytes)}
		return &s, nil
	})
}

// func serveAction(ctx context.Context, cmd *cli.Command) error {
// 	webFS, err := fs.Sub(dist, "ui/dist")
// 	if err != nil {
// 		return err
// 	}

// 	customArgs := []string{
// 		"--no-first-run",
// 		"--remote-allow-origins=*",
// 		"--user-data-dir=/tmp/situation",
// 		"--no-default-browser-check",
// 		"--safebrowsing-disable-auto-update",
// 	}
// 	ui, err := lorca.New("", "", 1280, 720, customArgs...)
// 	if err != nil {
// 		return err
// 	}
// 	defer ui.Close()
// 	ui.SetBounds(lorca.Bounds{Left: 0, Top: 0, WindowState: lorca.WindowStateFullscreen})

// 	// Bind Go function to be available in JS. Go function may be long-running and
// 	// blocking - in JS it's represented with a Promise.
// 	ui.Bind("add", func(a, b int) int { return a + b })

// 	// Call JS function from Go. Functions may be asynchronous, i.e. return promises
// 	n := ui.Eval(`Math.random()`).Float()
// 	fmt.Println(n)

// 	// Call JS that calls Go and so on and so on...
// 	m := ui.Eval(`add(2, 3)`).Int()
// 	fmt.Println(m)

// 	conn, err := net.Listen("tcp", "127.0.0.1:0")
// 	if err != nil {
// 		return err
// 	}
// 	defer conn.Close()
// 	go http.Serve(conn, http.FileServerFS(webFS))
// 	ui.Load(fmt.Sprintf("http://%s", conn.Addr()))
// 	// Wait for the browser window to be closed
// 	<-ui.Done()
// 	return nil
// }

// import (
// 	"context"
// 	"encoding/hex"
// 	"fmt"
// 	"net/http"
// 	"net/rpc"
// 	"strings"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/situation-sh/situation/pkg/backends/selfrpc"
// 	"github.com/situation-sh/situation/pkg/utils"
// 	"github.com/urfave/cli/v3"
// )

// func randomToken() string {
// 	return strings.ToUpper(hex.EncodeToString(utils.RandBytes(16)))
// }

// var serveCmd = cli.Command{
// 	Name:   "serve",
// 	Usage:  "Server mode: ready to receive data from other agents",
// 	Action: serveAction,
// 	Flags: []cli.Flag{
// 		&cli.StringFlag{
// 			Name:  "addr",
// 			Usage: "Address to listen on",
// 			Value: ":8080",
// 		},
// 		&cli.StringFlag{
// 			Name:  "secret",
// 			Usage: "Server secret to enroll clients",
// 			Value: randomToken(),
// 		},
// 	},
// }

// func serveAction(ctx context.Context, cmd *cli.Command) error {
// 	// rpcServer := rpc.NewServer()
// 	// api := selfrpc.API(cmd.String("secret"))

// 	// if err := rpcServer.RegisterName("API", api); err != nil {
// 	// 	return err
// 	// }

// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/enroll", func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != "POST" {
// 			http.Error(w, "Bad request", http.StatusBadRequest)
// 			return
// 		}
// 	})

// 	// RPC pat
// 	http.HandleFunc(rpc.DefaultRPCPath, func(w http.ResponseWriter, r *http.Request) {
// 		// Example header-based auth
// 		fmt.Printf("Incoming connection from %#+v\n", r)

// 		if r.Method != "CONNECT" {
// 			u, err := uuid.Parse(r.Header.Get("Authorization"))
// 			if err != nil {
// 				http.Error(w, "Bad request", http.StatusBadRequest)
// 				return
// 			}
// 			if api.Check(u) == false {
// 				http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 				return
// 			}
// 		}
// 		// Delegate to RPC server (it will Hijack and ServeConn)
// 		rpcServer.ServeHTTP(w, r)
// 	})

// 	// Optional: /debug/rpc endpoints can be added if you like, but protect them.

// 	server := &http.Server{
// 		Addr:              cmd.String("addr"),
// 		TLSConfig:         selfrpc.ServerTLSConfig(),
// 		ReadHeaderTimeout: time.Millisecond * 500,
// 	}

// 	fmt.Printf("Server listening on %s\n", server.Addr)
// 	if !cmd.IsSet("secret") {
// 		fmt.Println("Enroll token:", cmd.String("secret"))
// 	}
// 	return server.ListenAndServe()
// }
