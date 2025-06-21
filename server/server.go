package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gaesemo/tech-blog-api/go/service/auth/v1/authv1connect"
	authsvc "github.com/gaesemo/tech-blog-server/service/auth/v1"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

type app struct {
	port   string
	logger *slog.Logger
}

type Option func(*app)

func New(opts ...Option) *app {
	svr := &app{
		port: "8080",
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
		})),
	}

	for _, opt := range opts {
		opt(svr)
	}
	return svr
}
func WithPort(port string) Option {
	return func(a *app) {
		a.port = port
	}
}
func WithLogger(logger *slog.Logger) Option {
	return func(a *app) {
		a.logger = logger
	}
}

func (a *app) Serve(ctx context.Context) error {
	auth := authsvc.New(
		a.logger,
		authsvc.WithGitHubOAuth2(&oauth2.Config{
			ClientID:     os.Getenv("GITHUB_OAUTH2_CLIENT_ID"),
			ClientSecret: os.Getenv("GITHUB_OAUTH2_CLIENT_SECRET"),
			Endpoint:     endpoints.GitHub,
			RedirectURL:  os.Getenv("GITHUB_OAUTH2_DEFAULT_CALLBACK_URL"),
			Scopes:       []string{"user"}, // https://docs.github.com/ko/apps/oauth-apps/building-oauth-apps/scopes-for-oauth-apps#available-scopes
		}))
	path, handler := authv1connect.NewAuthServiceHandler(auth)

	mux := http.NewServeMux()
	mux.Handle(path, handler)

	server := &http.Server{
		Addr:    ":" + a.port,
		Handler: mux,
	}
	serverShutdownCallbacks := []func(){
		func() {
			slog.Info("gracefully shutting down x...")
		},
		func() {
			slog.Info("gracefully shutting down y...")
		},
		func() {
			slog.Info("gracefully shutting down z...")
		},
	}
	for _, callback := range serverShutdownCallbacks {
		server.RegisterOnShutdown(callback)
	}

	slog.InfoContext(ctx, "starting server...", slog.String("port", a.port))

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil {
			slog.ErrorContext(ctx, "server error", slog.Any("error", err))
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	slog.InfoContext(ctx, "context got cancelled")
	slog.InfoContext(ctx, "shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.ErrorContext(ctx, "gracefully shutting down server", slog.Any("error", err))
		return err
	}

	slog.InfoContext(ctx, "server stopped...")
	return nil
}
