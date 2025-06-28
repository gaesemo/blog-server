package v1

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"connectrpc.com/authn"
	"connectrpc.com/connect"
	postv1 "github.com/gaesemo/blog-api/go/service/post/v1"
	"github.com/gaesemo/blog-api/go/service/post/v1/postv1connect"
	typesv1 "github.com/gaesemo/blog-api/go/types/v1"
	"github.com/gaesemo/blog-server/gen/db/postgres"
	"github.com/gaesemo/blog-server/pkg/cursor"
	"github.com/gaesemo/blog-server/pkg/transaction"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ postv1connect.PostServiceHandler = (*service)(nil)

func New(
	logger *slog.Logger,
	db *pgx.Conn,
	timeNow func() time.Time,
) postv1connect.PostServiceHandler {
	return &service{
		logger:  logger,
		db:      db,
		queries: postgres.New(db),
		timeNow: timeNow,
	}
}

type service struct {
	logger  *slog.Logger
	db      *pgx.Conn
	queries *postgres.Queries
	timeNow func() time.Time
}

// Create implements postv1connect.PostServiceHandler.
func (s *service) Create(ctx context.Context, req *connect.Request[postv1.CreateRequest]) (*connect.Response[postv1.CreateResponse], error) {
	uid := authn.GetInfo(ctx).(*int64)
	if uid == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("author not found"))
	}
	content := req.Msg.PostContent
	title := content.Title
	body := content.Body

	type Result struct {
		User *postgres.User
		Post *postgres.Post
	}

	tx := transaction.New[Result](
		s.db,
		pgx.TxOptions{
			IsoLevel:   pgx.RepeatableRead,
			AccessMode: pgx.ReadWrite,
		},
		s.queries,
	)
	result, txErr := tx.Exec(ctx, func(c context.Context, q *postgres.Queries) (*Result, error) {
		user, err := q.GetUserById(c, *uid)
		if err != nil {
			return nil, fmt.Errorf("user not found: %v", err)
		}
		post, err := q.CreatePost(c, postgres.CreatePostParams{
			Likes:     0,
			Views:     0,
			Title:     title,
			Body:      body,
			UserID:    user.ID,
			CreatedAt: pgtype.Timestamptz{Time: s.timeNow(), Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: s.timeNow(), Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("insert new post: %v", err)
		}
		return &Result{
			User: &user,
			Post: &post,
		}, nil
	})
	if txErr != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("creating post: %v", txErr))
	}
	return connect.NewResponse(&postv1.CreateResponse{
		Post: &typesv1.Post{
			Id:    result.Post.ID,
			Likes: result.Post.Likes,
			Views: result.Post.Views,
			Author: &typesv1.User{
				Id:               result.User.ID,
				Username:         result.User.Username,
				Email:            result.User.Email,
				AvatarUrl:        result.User.AvatarUrl,
				AboutMe:          result.User.AboutMe,
				IdentityProvider: typesv1.IdentityProvider(typesv1.IdentityProvider_value[result.User.IdentityProvider]),
				CreatedAt:        timestamppb.New(result.User.CreatedAt.Time),
				UpdatedAt:        timestamppb.New(result.User.UpdatedAt.Time),
				DeletedAt:        nil,
			},
			Content:   content,
			CreatedAt: timestamppb.New(result.Post.CreatedAt.Time),
			UpdatedAt: timestamppb.New(result.Post.UpdatedAt.Time),
		},
	}), nil
}

// Delete implements postv1connect.PostServiceHandler.
func (s *service) Delete(ctx context.Context, req *connect.Request[postv1.DeleteRequest]) (*connect.Response[postv1.DeleteResponse], error) {
	uid := authn.GetInfo(ctx).(*int64)
	if uid == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("author not found"))
	}
	return nil, nil
}

// Detail implements postv1connect.PostServiceHandler.
func (s *service) Detail(ctx context.Context, req *connect.Request[postv1.DetailRequest]) (*connect.Response[postv1.DetailResponse], error) {
	type Result struct {
		User *postgres.User
		Post *postgres.Post
	}

	tx := transaction.New[Result](
		s.db,
		pgx.TxOptions{
			IsoLevel:   pgx.RepeatableRead,
			AccessMode: pgx.ReadWrite,
		},
		s.queries,
	)
	result, txErr := tx.Exec(ctx, func(c context.Context, q *postgres.Queries) (*Result, error) {
		post, err := q.GetPostById(ctx, req.Msg.Id)
		if err != nil {
			return nil, fmt.Errorf("post not found: %v", err)
		}
		user, err := q.GetUserById(c, post.UserID)
		if err != nil {
			return nil, fmt.Errorf("user not found: %v", err)
		}
		if err != nil {
			return nil, fmt.Errorf("insert new post: %v", err)
		}
		return &Result{
			User: &user,
			Post: &post,
		}, nil
	})
	if txErr != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("creating post: %v", txErr))
	}
	return connect.NewResponse(&postv1.DetailResponse{
		Post: &typesv1.Post{
			Id:    result.Post.ID,
			Likes: result.Post.Likes,
			Views: result.Post.Views,
			Author: &typesv1.User{
				Id:               result.User.ID,
				Username:         result.User.Username,
				Email:            result.User.Email,
				AvatarUrl:        result.User.AvatarUrl,
				AboutMe:          result.User.AboutMe,
				IdentityProvider: typesv1.IdentityProvider(typesv1.IdentityProvider_value[result.User.IdentityProvider]),
				CreatedAt:        timestamppb.New(result.User.CreatedAt.Time),
				UpdatedAt:        timestamppb.New(result.User.UpdatedAt.Time),
				DeletedAt:        nil,
			},
			Content: &typesv1.PostContent{
				Title: result.Post.Title,
				Body:  result.Post.Body,
			},
			CreatedAt: timestamppb.New(result.Post.CreatedAt.Time),
			UpdatedAt: timestamppb.New(result.Post.UpdatedAt.Time),
		},
	}), nil
}

// List implements postv1connect.PostServiceHandler.
func (s *service) List(ctx context.Context, req *connect.Request[postv1.ListRequest]) (*connect.Response[postv1.ListResponse], error) {
	cur := cursor.MustParseInt64(req.Msg.Cursor)

	rows, err := s.queries.ListRecentPosts(ctx, postgres.ListRecentPostsParams{
		Limit:  10,
		Cursor: cur,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("retreiving posts: %v", err))
	}

	last := len(rows)
	posts := []*typesv1.Post{}
	var c int64
	for i, p := range rows {
		posts = append(posts, pbPost(&p))
		if i == last-1 {
			c = p.ID
		}
	}

	return connect.NewResponse(&postv1.ListResponse{
		Posts: posts,
		Next:  cursor.FromInt64(c),
	}), nil
}

// Update implements postv1connect.PostServiceHandler.
func (s *service) Update(ctx context.Context, req *connect.Request[postv1.UpdateRequest]) (*connect.Response[postv1.UpdateResponse], error) {
	uid := authn.GetInfo(ctx).(*int64)
	if uid == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("author not found"))
	}
	return nil, nil
}

func pbPost(p *postgres.Post) *typesv1.Post {
	return &typesv1.Post{
		Id:        p.ID,
		Likes:     p.Likes,
		Views:     p.Views,
		Author:    &typesv1.User{},
		Content:   &typesv1.PostContent{},
		CreatedAt: &timestamppb.Timestamp{},
		UpdatedAt: &timestamppb.Timestamp{},
		DeletedAt: &timestamppb.Timestamp{},
	}
}

func pbUser(u *postgres.User) *typesv1.User {
	var deletedAt *timestamppb.Timestamp
	if u.DeletedAt.Valid {
		deletedAt = timestamppb.New(u.DeletedAt.Time)
	}
	return &typesv1.User{
		Id:               u.ID,
		Username:         u.Username,
		Email:            u.Email,
		AvatarUrl:        u.AvatarUrl,
		AboutMe:          u.AboutMe,
		IdentityProvider: pbIdentityProvider(u.IdentityProvider),
		CreatedAt:        timestamppb.New(u.CreatedAt.Time),
		UpdatedAt:        timestamppb.New(u.UpdatedAt.Time),
		DeletedAt:        deletedAt,
	}
}

func pbIdentityProvider(ipd string) typesv1.IdentityProvider {
	return typesv1.IdentityProvider(typesv1.IdentityProvider_value[ipd])
}
