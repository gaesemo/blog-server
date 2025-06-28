package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"connectrpc.com/authn"
	"connectrpc.com/connect"
	connectcors "connectrpc.com/cors"
	"github.com/gaesemo/blog-api/go/service/auth/v1/authv1connect"
	"github.com/gaesemo/blog-api/go/service/post/v1/postv1connect"
	"github.com/gaesemo/blog-server/pkg/middleware"
	"github.com/gaesemo/blog-server/pkg/oauth"
	authsvc "github.com/gaesemo/blog-server/service/auth/v1"
	postsvc "github.com/gaesemo/blog-server/service/post/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/cors"
	"golang.org/x/sync/errgroup"
)

// interface - type - const - var - new func - public receiver func - private receiver func - public func - private func

type Server struct {
	logger *slog.Logger
	port   uint16
	db     *pgx.Conn
	// objstorage
}

func New(logger *slog.Logger, port uint16, db *pgx.Conn) *Server {
	return &Server{
		logger: logger,
		port:   port,
		db:     db,
	}
}

func (s *Server) Serve(ctx context.Context) error {
	timeNow := func() time.Time {
		return time.Now().UTC()
	}
	randStr := func() string {
		return uuid.NewString()
	}

	db := s.db
	httpClient := &http.Client{Timeout: 10 * time.Second}
	authService := authsvc.New(
		slog.Default(),
		httpClient,
		db,
		timeNow,
		randStr,
		authsvc.WithGitHubOAuthApp(oauth.NewGitHub(httpClient, randStr)),
	)
	postService := postsvc.New(
		slog.Default(),
		db,
		timeNow,
	)

	mux := http.NewServeMux()

	authorizer := authn.NewMiddleware(
		middleware.Authorize,
	)
	{
		path, svcHandler := authv1connect.NewAuthServiceHandler(
			authService,
			connect.WithInterceptors(middleware.UnaryLogger()),
		) // TOOD: add request id interceptor, add logging interceptor,
		mux.Handle(path, svcHandler)
	}
	{
		path, svcHandler := postv1connect.NewPostServiceHandler(
			postService,
			connect.WithInterceptors(middleware.UnaryLogger()),
		)
		svcHandler = authorizer.Wrap(svcHandler)
		mux.Handle(path, svcHandler)
	}

	handler := withCORS(mux)

	addr := ":" + strconv.FormatUint(uint64(s.port), 10)
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
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

func withCORS(h http.Handler) http.Handler {
	allowedHeaders := connectcors.AllowedHeaders()
	allowedHeaders = append(allowedHeaders, "Credentials")
	middlewares := cors.New(cors.Options{
		AllowedOrigins:       []string{"http://localhost:3000"},
		AllowedMethods:       connectcors.AllowedMethods(),
		AllowedHeaders:       allowedHeaders,
		AllowCredentials:     true,
		ExposedHeaders:       connectcors.ExposedHeaders(),
		Debug:                true,
		OptionsSuccessStatus: http.StatusOK,
	})
	return middlewares.Handler(h)
}
