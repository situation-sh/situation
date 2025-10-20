package cmd

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
