package v1

import (
	"context"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	postv1 "github.com/gaesemo/blog-api/go/service/post/v1"
	"github.com/gaesemo/blog-api/go/service/post/v1/postv1connect"
	"github.com/jackc/pgx/v5"
)

var _ postv1connect.PostServiceHandler = (*service)(nil)

func New(
	logger *slog.Logger,
	db *pgx.Conn,
	timeNow func() time.Time,
) postv1connect.PostServiceHandler {
	return &service{}
}

type service struct {
}

// Create implements postv1connect.PostServiceHandler.
func (s *service) Create(context.Context, *connect.Request[postv1.CreateRequest]) (*connect.Response[postv1.CreateResponse], error) {
	panic("unimplemented")
}

// Delete implements postv1connect.PostServiceHandler.
func (s *service) Delete(context.Context, *connect.Request[postv1.DeleteRequest]) (*connect.Response[postv1.DeleteResponse], error) {
	panic("unimplemented")
}

// Detail implements postv1connect.PostServiceHandler.
func (s *service) Detail(context.Context, *connect.Request[postv1.DetailRequest]) (*connect.Response[postv1.DetailResponse], error) {
	panic("unimplemented")
}

// List implements postv1connect.PostServiceHandler.
func (s *service) List(context.Context, *connect.Request[postv1.ListRequest]) (*connect.Response[postv1.ListResponse], error) {
	panic("unimplemented")
}

// Update implements postv1connect.PostServiceHandler.
func (s *service) Update(context.Context, *connect.Request[postv1.UpdateRequest]) (*connect.Response[postv1.UpdateResponse], error) {
	panic("unimplemented")
}
