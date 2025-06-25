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

	"github.com/gaesemo/tech-blog-api/go/service/auth/v1/authv1connect"
	"github.com/gaesemo/tech-blog-server/pkg/oauth"
	authsvc "github.com/gaesemo/tech-blog-server/service/auth/v1"
	"github.com/jackc/pgx/v5"
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
		return time.Time{}
	}
	randStr := func() string {
		return "some-randomized-string"
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

	path, handler := authv1connect.NewAuthServiceHandler(auth) // TOOD: add request id interceptor, add logging interceptor,

	mux := http.NewServeMux()
	mux.Handle(path, handler) // TODO: add middlewares e.g. panic recoverer, request logger
	mux.HandleFunc("GET /oauth/github/callback", func(w http.ResponseWriter, r *http.Request) {
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
