package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	pb "github.com/Zartex-the-art/sei-ratelimiter/api/proto"
	grpcserver "github.com/Zartex-the-art/sei-ratelimiter/internal/grpc"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/algorithms"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/config"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/handlers"
	"github.com/Zartex-the-art/sei-ratelimiter/internal/store"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()

	rs := store.NewRedisStore(cfg.RedisURL)

	if err := rs.Ping(context.Background()); err != nil {
		log.Printf("WARNING: Redis not reachable at %s: %v", cfg.RedisURL, err)
		log.Printf("WARNING: Server starting in degraded mode — all /check calls will return 503")
		log.Printf("WARNING: Server will auto-recover when Redis becomes available")
	} else {
		log.Printf("Redis connected at %s", cfg.RedisURL)
	}

	// Sanity check: all algorithms register correctly at startup
	for _, algo := range algorithms.ValidAlgorithms() {
		if _, err := algorithms.NewLimiter(algo, rs, 100, 60); err != nil {
			log.Fatalf("factory sanity check failed for %s: %v", algo, err)
		}
	}

	log.Printf(
		"Algorithm factory: all %d algorithms registered",
		len(algorithms.ValidAlgorithms()),
	)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","node":%q}`, cfg.NodeID)
	})
	http.HandleFunc("POST /check", handlers.CheckHandler(rs, rs.Client()))
	http.HandleFunc("GET /rules/{id}", handlers.GetRuleHandler(rs.Client()))
	http.HandleFunc("DELETE /rules/{id}", handlers.DeleteRuleHandler(rs.Client()))

	http.HandleFunc("/rules", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.CreateRuleHandler(rs.Client())(w, r)

		case http.MethodGet:
			handlers.ListRulesHandler(rs.Client())(w, r)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	log.Printf(
		"starting sei-ratelimiter node=%s port=%s",
		cfg.NodeID,
		cfg.Port,
	)
	grpcLis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen on grpc port: %v", err)
	}

	grpcSrv := grpc.NewServer()

	pb.RegisterRateLimiterServer(
		grpcSrv,
		&grpcserver.Server{
			Store: rs,
			Redis: rs.Client(),
		},
	)

	go func() {
		log.Printf("gRPC server listening on %s", cfg.GRPCPort)

		if err := grpcSrv.Serve(grpcLis); err != nil {
			log.Fatalf("grpc server error: %v", err)
		}
	}()

	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}

}
