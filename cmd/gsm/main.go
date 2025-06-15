package main

import (
	"log/slog"
	"os"

	"github.com/gaesemo/tech-blog-api/go/service/auth/v1/authv1connect"
	authsvc "github.com/gaesemo/tech-blog-server/service/auth/v1"
)

func main() {
	//
	// db connection
	//

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}))
	auth := authsvc.New(logger)
	authv1connect.NewAuthServiceHandler(auth)
}
