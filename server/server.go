package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gaesemo/tech-blog-api/go/service/auth/v1/authv1connect"
	"github.com/gaesemo/tech-blog-server/gen/db/postgres"
	"github.com/gaesemo/tech-blog-server/pkg/oauthapp"
	authsvc "github.com/gaesemo/tech-blog-server/service/auth/v1"
	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/errgroup"
)

// interface - type - const - var - new func - public receiver func - private receiver func - public func - private func

type Server struct {
	port uint16
	db   *pgx.Conn
	// objstorage
}

func New(logger *slog.Logger, port uint16, db *pgx.Conn) *Server {
	return &Server{
		port: port,
		db:   db,
	}
}

func (s *Server) Serve(ctx context.Context) error {
	timeNow := func() time.Time {
		return time.Time{}
	}
	randStr := func() string {
		return "some-randomized-string"
	}

	db := s.db
	queries := postgres.New(db)
	httpClient := &http.Client{Timeout: 10 * time.Second}

	auth := authsvc.New(
		slog.Default(),
		db,
		queries,
		httpClient,
		timeNow, // timeNow
		randStr, // randStr
		authsvc.WithGitHubOAuthApp(oauthapp.NewGitHub(httpClient, randStr)),
	)

	path, handler := authv1connect.NewAuthServiceHandler(auth) // TOOD: add request id interceptor, add logging interceptor,

	mux := http.NewServeMux()
	mux.Handle(path, handler) // TODO: add middlewares e.g. panic recoverer, request logger

	addr := ":" + strconv.FormatUint(uint64(s.port), 10)
	server := &http.Server{
		Addr:    addr, // :port
		Handler: mux,
	}

	eg, ctx := errgroup.WithContext(ctx)
	onShutdown := func() error {
		<-ctx.Done()
		err := server.Shutdown(ctx)
		if err != nil {
			return fmt.Errorf("shutting down http server: %v", err)
		}
		return nil
	}
	eg.Go(onShutdown)

	serve := func() error {
		slog.InfoContext(ctx, "start server", slog.String("address", server.Addr))
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
	eg.Go(serve)

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("server stopped: %v", err)
	}
	return nil
}
