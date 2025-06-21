package cmd

import (
	"log/slog"
	"os"

	"github.com/gaesemo/tech-blog-server/server"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:  "serve",
	RunE: serve,
	Example: `
	serve -p=8081 -ll=debug
	serve --port=8081 --log-level=debug
`,
}

// flag value
var (
	port string
	ll   string
)

func serve(cmd *cobra.Command, args []string) error {

	ctx := cmd.Context()

	level := slog.LevelInfo
	level.UnmarshalText([]byte(ll))

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}))
	slog.SetDefault(logger)

	srv := server.New(server.WithPort(port), server.WithLogger(logger))
	if err := srv.Serve(ctx); err != nil {
		return err
	}

	return nil
}

func init() {
	serveCmd.Flags().StringVarP(&port, "port", "p", "8080", "set request listening port")
	serveCmd.Flags().StringVar(&ll, "log-level", "info", "set log level. value must be one of the [debug, info, warn, error]")
}
