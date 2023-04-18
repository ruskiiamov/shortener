// Package grpcserver implements structure to handle gRPC requests.
package grpcserver

import (
	"context"
	"errors"
	"time"

	pb "github.com/ruskiiamov/shortener/internal/proto"
	"github.com/ruskiiamov/shortener/internal/url"
	"github.com/ruskiiamov/shortener/internal/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const userIDctxKey ctxKey = "user_id"
const authHeader = "auth"

type ctxKey string

type grpcServer struct {
	pb.UnimplementedShortenerServer
	urlConverter url.Converter
	delBuf       chan *url.DelBatch
}

// NewGRPCServer returns gRPC server implementation.
func NewGRPCServer(u url.Converter, delBuf chan *url.DelBatch) *grpcServer {
	return &grpcServer{
		urlConverter: u,
		delBuf:       delBuf,
	}
}

// NewAuthInterceptor returns interceptor for auth.
func NewAuthInterceptor(ua user.Authorizer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var token string

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			values := md.Get(authHeader)
			if len(values) > 0 {
				token = values[0]
			}
		}

		userID, err := ua.GetUserID(token)
		if err != nil {
			userID, token, err = ua.CreateUser()
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			err = grpc.SetHeader(ctx, metadata.Pairs(authHeader, token))
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		}

		ctxAuth := context.WithValue(ctx, userIDctxKey, userID)

		return handler(ctxAuth, req)
	}
}

// GetURL implements interface of getting URL.
func (g *grpcServer) GetURL(ctx context.Context, in *pb.GetURLRequest) (*pb.GetURLResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	shortURL, err := g.urlConverter.GetOriginal(ctx, in.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &pb.GetURLResponse{Url: shortURL.Original}, nil
}

// AddURL implements interface of saving new URL.
func (g *grpcServer) AddURL(ctx context.Context, in *pb.AddURLRequest) (*pb.AddURLResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	userID, ok := ctx.Value(userIDctxKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Internal, "User ID error")
	}

	var errDupl *url.ErrURLDuplicate

	url, err := g.urlConverter.Shorten(ctx, userID, in.Url)
	if errors.As(err, &errDupl) {
		return nil, status.Error(codes.AlreadyExists, "URL ID is "+errDupl.EncodedID)
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AddURLResponse{Id: url.EncodedID}, nil
}

// AddURLBatch implements interface of saving URL batch.
func (g *grpcServer) AddURLBatch(ctx context.Context, in *pb.AddURLBatchRequest) (*pb.AddURLBatchResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	userID, ok := ctx.Value(userIDctxKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Internal, "User ID error")
	}

	var originals []string
	for _, item := range in.Urls {
		originals = append(originals, item.Url)
	}
	shortURLs, err := g.urlConverter.ShortenBatch(ctx, userID, originals)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var ids []*pb.AddURLBatchResponseItem
	for _, item := range in.Urls {
		for _, shortURL := range shortURLs {
			if shortURL.Original == item.Url {
				ids = append(ids, &pb.AddURLBatchResponseItem{
					CorrelationId: item.CorrelationId,
					Id:            shortURL.EncodedID,
				})
				break
			}
		}
	}

	if len(in.Urls) != len(ids) {
		return nil, status.Error(codes.Internal, "Shorten Batch error")
	}

	return &pb.AddURLBatchResponse{Ids: ids}, nil
}

// GetAllURL implements interface of getting all URL by user.
func (g *grpcServer) GetAllURL(ctx context.Context, in *pb.GetAllURLRequest) (*pb.GetAllURLResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	userID, ok := ctx.Value(userIDctxKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Internal, "User ID error")
	}

	shortURLs, err := g.urlConverter.GetAllByUser(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var urls []*pb.GetAllURLResponseItem
	for _, shortURL := range shortURLs {
		urls = append(urls, &pb.GetAllURLResponseItem{
			Id:  shortURL.EncodedID,
			Url: shortURL.Original,
		})
	}

	return &pb.GetAllURLResponse{Urls: urls}, nil
}

// DeleteURLBatch implements interface of deleting URL batch.
func (g *grpcServer) DeleteURLBatch(ctx context.Context, in *pb.DeleteURLBatchRequest) (*pb.DeleteURLBatchResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	userID, ok := ctx.Value(userIDctxKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Internal, "User ID error")
	}

	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Internal, "Context canceled")
	default:
		g.delBuf <- &url.DelBatch{
			UserID:     userID,
			EncodedIDs: in.Ids,
		}
	}

	return &pb.DeleteURLBatchResponse{}, nil
}

// GetStats implements interface of getting service statistics.
func (g *grpcServer) GetStats(ctx context.Context, in *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	urls, users, err := g.urlConverter.GetStats(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetStatsResponse{
		Urls:  int32(urls),
		Users: int32(users),
	}, nil
}

// PingDB implements interface of the databease ping.
func (g *grpcServer) PingDB(ctx context.Context, in *pb.PingDBRequest) (*pb.PingDBResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err := g.urlConverter.PingKeeper(ctx); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.PingDBResponse{}, nil
}
