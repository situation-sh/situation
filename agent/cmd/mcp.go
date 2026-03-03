package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/situation-sh/situation/agent/config"
	"github.com/situation-sh/situation/pkg/store"
	"github.com/situation-sh/situation/pkg/utils"
	"github.com/uptrace/bun/dialect"
	"github.com/urfave/cli/v3"
)

// variables filled by the linker
var (
	POSTGRES_SVG  string = ""
	SQLITE_SVG    string = ""
	SITUATION_SVG string = ""
	SQL_SVG       string = ""
)

var (
	mcpTransport       string = "stdio"
	mcpValidTransports        = []string{"stdio"} // []string{"http", "stdio"}
	// mcpPort            uint16 = 8080
	// mcpHost            string = "localhost"
)

// currently we do not implement the HTTP transport because of security
// and binary weight concerns but we keep the code and flags for future
// use and to avoid breaking changes
var mcpCmd = cli.Command{
	Name:   "mcp",
	Usage:  "Start an MCP server to query collected data",
	Action: mcpAction,
	Flags: []cli.Flag{
		// &cli.StringFlag{
		// 	Name:        "host",
		// 	Value:       "localhost",
		// 	Destination: &mcpHost,
		// 	Usage:       "Host to listen on for MCP server",
		// },
		// &cli.Uint16Flag{
		// 	Name:        "port",
		// 	Value:       8080,
		// 	Destination: &mcpPort,
		// 	Usage:       "Port to listen on for MCP server",
		// },
		&cli.StringFlag{
			Name:        "transport",
			Aliases:     []string{"t"},
			Value:       mcpValidTransports[0],
			Destination: &mcpTransport,
			Validator: func(s string) error {
				if !utils.Includes(mcpValidTransports, s) {
					return fmt.Errorf("invalid transport: %s (choose from %v)", s, mcpValidTransports)
				}
				return nil
			},
		},
	},
}

func init() {
	defineDB()
	flags, err := config.SomeFlags("db")
	if err != nil {
		panic(err)
	}
	mcpCmd.Flags = append(mcpCmd.Flags, flags...)
}

var server = mcp.NewServer(
	&mcp.Implementation{
		Name:       "situation",
		Title:      "Situation MCP server",
		Version:    config.Version,
		WebsiteURL: "https://github.com/situation-sh/situation",
		Icons: []mcp.Icon{
			{
				Source:   SITUATION_SVG,
				Sizes:    []string{"any"},
				MIMEType: "image/svg+xml",
			},
		},
	},
	&mcp.ServerOptions{
		Logger: slog.New(newSlogHandler()),
	},
)

func mcpError(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)}},
	}
}

type QueryArgs struct {
	SQL string `json:"sql" jsonschema:"description:SQL query to execute"`
}

func mcpAction(ctx context.Context, cmd *cli.Command) error {
	storage, err := store.NewStorage(db,
		store.WithAgent(config.AgentString()),
		store.WithErrorHandler(func(err error) {
			logger.WithField("on", "storage").Warn(err)
		}),
		store.ReadOnly(),
	)
	if err != nil {
		return fmt.Errorf("failed to create storage: %v", err)
	}

	schema, err := storage.RawSchema()
	if err != nil {
		return fmt.Errorf("failed to get raw schema: %w", err)
	}

	// expose the raw SQL schema as an MCP resource
	server.AddResource(
		&mcp.Resource{
			Description: "Database SQL schema",
			Name:        "schema",
			Title:       "Situation schema",
			MIMEType:    "text/plain",
			URI:         fmt.Sprintf("file:///situation/%s/schema.sql", config.Version),
			Size:        int64(len([]byte(schema))),
			Icons: []mcp.Icon{
				{
					Source:   SQL_SVG,
					Sizes:    []string{"any"},
					MIMEType: "image/svg+xml",
				},
			},
		},
		func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

			if err != nil {
				return nil, fmt.Errorf("failed to get raw schema: %w", err)
			}
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{
					{URI: req.Params.URI, Text: schema, MIMEType: "text/plain"}},
			}, nil
		},
	)

	destructive := false
	openWorld := false
	queryTool := mcp.Tool{
		Title:       "Run SQL query on your infrastructure data",
		Name:        "query",
		Description: "Execute a read-only SQL query, returns JSON rows",
		Meta:        mcp.Meta{"dialect": storage.Dialect().String()},
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: &destructive,
			IdempotentHint:  true,
			OpenWorldHint:   &openWorld,
			ReadOnlyHint:    true,
		},
	}
	if storage.Dialect() == dialect.SQLite {
		queryTool.Icons = append(queryTool.Icons,
			mcp.Icon{
				Source:   SQLITE_SVG,
				Sizes:    []string{"any"},
				MIMEType: "image/svg+xml",
			},
		)
	} else if storage.Dialect() == dialect.PG {
		queryTool.Icons = append(queryTool.Icons,
			mcp.Icon{
				Source:   POSTGRES_SVG,
				Sizes:    []string{"any"},
				MIMEType: "image/svg+xml",
			},
		)
	}

	mcp.AddTool(server, &queryTool, func(ctx context.Context, req *mcp.CallToolRequest, args QueryArgs) (*mcp.CallToolResult, any, error) {
		var results []map[string]any
		rows, err := storage.DB().QueryContext(ctx, args.SQL)
		if err != nil {
			return mcpError(err), nil, nil
		}
		defer rows.Close()

		cols, _ := rows.Columns()
		for rows.Next() {
			vals := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			rows.Scan(ptrs...)
			row := make(map[string]any, len(cols))
			for i, c := range cols {
				row[c] = vals[i]
			}
			results = append(results, row)
		}

		out, _ := json.MarshalIndent(results, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(out)}},
		}, nil, nil
	})

	switch mcpTransport {
	case "http":
		return nil
		// handler := mcp.NewStreamableHTTPHandler(
		// 	func(*http.Request) *mcp.Server { return server },
		// 	&mcp.StreamableHTTPOptions{},
		// )
		// return http.ListenAndServe(
		// 	net.JoinHostPort(mcpHost, fmt.Sprintf("%d", mcpPort)),
		// 	handler,
		// )
	case "stdio":
		return server.Run(context.Background(), &mcp.StdioTransport{})
	default:
		return fmt.Errorf("unsupported transport: %s", mcpTransport)
	}
}
