package middleware

import (
	"context"
	"log/slog"

	"connectrpc.com/connect"
)

func UnaryLogger() connect.UnaryInterceptorFunc {
	return logUnary
}

func logUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		procedure := req.Spec().Procedure
		slog.InfoContext(ctx, "request", slog.String("calling", procedure), slog.String("method", req.HTTPMethod()), slog.Any("headers", req.Header()), slog.Any("body", req.Any()))
		resp, err := next(ctx, req)
		if err != nil {
			slog.ErrorContext(ctx, "response", slog.String("calling", procedure), slog.Any("error", err))
		} else {
			slog.InfoContext(ctx, "response", slog.String("calling", procedure), slog.Any("headers", resp.Header()), slog.Any("body", resp.Any()))
		}
		return resp, nil
	}
}
