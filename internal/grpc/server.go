package grpc

import (
	"context"

	pb "github.com/Zartex-the-art/sei-ratelimiter/api/proto"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/algorithms"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/models"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/store"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	pb.UnimplementedRateLimiterServer
	Store store.Store
	Redis *redis.Client
}

func (s *Server) Check(
	ctx context.Context,
	req *pb.CheckRequest,
) (*pb.CheckResponse, error) {

	checkReq := models.CheckRequest{
		ClientID:   req.ClientId,
		Algorithm:  req.Algorithm,
		Limit:      int(req.Limit),
		WindowSecs: int(req.WindowSecs),
	}

	algo := checkReq.Algorithm
	limit := checkReq.Limit
	windowSecs := checkReq.WindowSecs

	limiter, err := algorithms.NewLimiter(
		algo,
		s.Store,
		limit,
		windowSecs,
	)
	if err != nil {
		return &pb.CheckResponse{
			Error: err.Error(),
		}, nil
	}

	allowed, remaining, err :=
		limiter.Allow(ctx, checkReq.ClientID)

	if err != nil {
		return &pb.CheckResponse{
			Error: err.Error(),
		}, nil
	}

	return &pb.CheckResponse{
		Allowed:   allowed,
		Remaining: int32(remaining),
		Algorithm: algo,
		ClientId:  checkReq.ClientID,
	}, nil
}
