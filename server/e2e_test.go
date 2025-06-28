package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	authv1 "github.com/gaesemo/blog-api/go/service/auth/v1"
	typesv1 "github.com/gaesemo/blog-api/go/types/v1"
	"github.com/gaesemo/blog-server/client"
	"github.com/gaesemo/blog-server/config"
	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"golang.org/x/sync/errgroup"
)

type cleanUpFunc func(c context.Context) error

func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute) // e2e test should be done in 3 minutes
	defer cancel()

	db, cleanUpFuncs, err := setUp(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "setting up tests", slog.Any("error", err))
		return
	}

	eg, ctx := errgroup.WithContext(ctx)

	cleanUpOnCancel := func() error {
		for _, f := range cleanUpFuncs {
			f(ctx)
		}
		return nil
	}
	eg.Go(cleanUpOnCancel)

	srv := New(slog.Default(), 8080, db)
	serve := func() error {
		return srv.Serve(ctx)
	}
	eg.Go(serve)

	slog.InfoContext(ctx, "test set up")

	code := m.Run()

	slog.Info("clean up test")
	cancel()

	if err := eg.Wait(); err != nil {
		slog.Error("cleaning up test", slog.Any("error", err))
	}

	os.Exit(code)
}

func TestGetAuthUrl(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	c := client.Default()
	resp, err := c.Auth.GetAuthURL(ctx, &connect.Request[authv1.GetAuthURLRequest]{
		Msg: &authv1.GetAuthURLRequest{
			IdentityProvider: typesv1.IdentityProvider_IDENTITY_PROVIDER_GITHUB,
		},
	})
	require.NoError(t, err)
	t.Log(resp.Msg.AuthUrl)
}

func setUp(ctx context.Context) (*pgx.Conn, []cleanUpFunc, error) {
	if err := config.Load(); err != nil {
		return nil, nil, fmt.Errorf("loading config: %v", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var err error

	pgc, err := tcpg.Run(ctx, "postgres:17-alpine",
		tcpg.WithUsername(viper.GetString("RDB_USER")),
		tcpg.WithPassword(viper.GetString("RDB_PASSWORD")),
		tcpg.WithDatabase(viper.GetString("RDB_DATABASE")),
		tcpg.WithInitScripts("../db/postgres/schema.sql"),
		// TODO: initial schema set up
		// TODO: test data setup
		tcpg.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("running pg container: %v", err)
	}
	defer func() {
		if err != nil {
			pgc.Terminate(ctx)
		}
	}()

	connStr, err := pgc.ConnectionString(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("getting connection string: %v", err)
	}

	db, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, nil, fmt.Errorf("connecting db: %v", err)
	}
	defer func() {
		if err != nil {
			db.Close(ctx)
		}
	}()

	cleanUpFuncs := []cleanUpFunc{
		func(c context.Context) error {
			<-c.Done()
			return db.Close(c)
		},
		func(c context.Context) error {
			<-c.Done()
			return pgc.Terminate(c)
		},
	}

	return db, cleanUpFuncs, nil
}
