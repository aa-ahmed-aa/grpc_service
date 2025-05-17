// main.go
package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	"zendesk_grpc_service/internal/common"
	"zendesk_grpc_service/internal/rating"
	pb "zendesk_grpc_service/proto/ratingService/v1"
)

// runGRPCServer starts the gRPC server on the given address.
func runGRPCServer(dbPath, address string) error {
	db, err := common.OpenDB(dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	repo := &rating.RatingRepository{DB: db}
	service := &rating.RatingService{Repo: repo}

	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	pb.RegisterRatingServiceServer(grpcServer, service)
	log.Printf("gRPC server listening on %s", address)
	return grpcServer.Serve(lis)
}

func main() {
	if err := runGRPCServer("./database.db", ":50051"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
