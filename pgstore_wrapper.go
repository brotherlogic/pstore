package main

import (
	"context"
	"fmt"

	pb "github.com/brotherlogic/pstore/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func getPGStore() (*pgstore_wrapper, error) {
	conn, err := grpc.Dial("rstore.rstore:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(100*1024*1024)))
	if err != nil {
		return nil, fmt.Errorf("dial error on %v -> %w", "pstore.pstore:8080", err)
	}
	return &pgstore_wrapper{client: pb.NewPStoreServiceClient(conn)}, nil
}

type pgstore_wrapper struct {
	client pb.PStoreServiceClient
}

func (p *pgstore_wrapper) Name() string {
	return "pgstore"
}

func (p *pgstore_wrapper) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	return p.client.Read(ctx, req)
}
func (p *pgstore_wrapper) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	return p.client.Write(ctx, req)
}
func (p *pgstore_wrapper) Count(ctx context.Context, req *pb.CountRequest) (*pb.CountResponse, error) {
	return p.client.Count(ctx, req)
}
func (p *pgstore_wrapper) GetKeys(ctx context.Context, req *pb.GetKeysRequest) (*pb.GetKeysResponse, error) {
	return p.client.GetKeys(ctx, req)
}
func (p *pgstore_wrapper) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	return p.client.Delete(ctx, req)
}
