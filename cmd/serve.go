package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/gaesemo/blog-server/config"
	"github.com/gaesemo/blog-server/server"
	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	serveCmd.Flags().Uint16P("port", "p", uint16(8080), "set request listening port")
	viper.BindPFlag("port", serveCmd.Flags().Lookup("port"))

	serveCmd.Flags().String("log-level", "info", "set log level. value must be one of the [debug, info, warn, error]")
	viper.BindPFlag("log-level", serveCmd.Flags().Lookup("log-level"))
}

var serveCmd = &cobra.Command{
	Use:     "serve",
	PreRunE: loadConfig,
	RunE:    serve,
	Example: `
	serve -p=8080 --log-level=debug
`,
}

func loadConfig(cmd *cobra.Command, args []string) error {
	err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %v", err)
	}
	return nil
}

func serve(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	level := slog.LevelInfo
	level.UnmarshalText([]byte(viper.GetString("log-level")))

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})))
	port := viper.GetUint16("port")

	connstr := pgConnStr()
	pg, err := pgx.Connect(ctx, connstr)
	if err != nil {
		slog.ErrorContext(ctx, "connecting db: %v", slog.Any("error", err))
		return err
	}
	defer func() {
		if err := pg.Close(ctx); err != nil {
			slog.ErrorContext(ctx, "closing db: %v", slog.Any("error", err))
		}
	}()

	srv := server.New(slog.Default(), port, pg)
	if err := srv.Serve(ctx); err != nil {
		return err
	}

	return nil
}

func pgConnStr() string {
	user := viper.GetString("RDB_USER")
	password := viper.GetString("RDB_PASSWORD")
	host := viper.GetString("RDB_HOST")
	port := viper.GetUint16("RDB_PORT")
	database := viper.GetString("RDB_DATABASE")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, password, host, port, database)
	return dsn
}
