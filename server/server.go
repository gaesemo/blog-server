package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	connectcors "connectrpc.com/cors"
	"github.com/gaesemo/blog-api/go/service/auth/v1/authv1connect"
	"github.com/gaesemo/blog-server/pkg/oauth"
	authsvc "github.com/gaesemo/blog-server/service/auth/v1"
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

	auth := authsvc.New(
		slog.Default(),
		httpClient,
		db,
		timeNow,
		randStr,
		authsvc.WithGitHubOAuthApp(oauth.NewGitHub(httpClient, randStr)),
	)

	mux := http.NewServeMux()
	{
		path, handler := authv1connect.NewAuthServiceHandler(auth) // TOOD: add request id interceptor, add logging interceptor,
		mux.Handle(path, handler)
	}
	{
		var handler http.Handler
		path, handler := "/oauth/github/callback", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			code := r.URL.Query().Get("code")
			redirectURL := r.URL.Query().Get("redirect_uri")
			authToken, err := auth.GitHubCallback(ctx, code, redirectURL)
			if err != nil {
				params := url.Values{}
				params.Add("status", "error")
				params.Add("message", err.Error())
				http.Redirect(w, r, redirectURL+"?"+params.Encode(), http.StatusMovedPermanently)
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:     "auth_token",
				Value:    authToken,
				Path:     "/",
				MaxAge:   3600,
				HttpOnly: true,
			})
			http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
		})
		mux.Handle(path, handler)
	}

	addr := ":" + strconv.FormatUint(uint64(s.port), 10)
	server := &http.Server{
		Addr:    addr,
		Handler: withCORS(mux),
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
	middlewares := cors.New(cors.Options{
		AllowedOrigins:       []string{"http://localhost:3000"},
		AllowedMethods:       connectcors.AllowedMethods(),
		AllowedHeaders:       connectcors.AllowedHeaders(),
		ExposedHeaders:       connectcors.ExposedHeaders(),
		Debug:                true,
		OptionsSuccessStatus: http.StatusOK,
	})
	return middlewares.Handler(h)
}
