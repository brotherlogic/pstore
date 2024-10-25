package pstore_client

import (
	"context"
	"fmt"

	pb "github.com/brotherlogic/pstore/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PStoreClient interface {
	Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error)
	Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error)
	GetKeys(ctx context.Context, req *pb.GetKeysRequest) (*pb.GetKeysResponse, error)
	Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error)
	Count(ctx context.Context, req *pb.CountRequest) (*pb.CountResponse, error)
}

type pClient struct {
	pClient pb.PStoreServiceClient
}

func GetClient() (PStoreClient, error) {
	conn, err := grpc.Dial("pstore.pstore:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(100*1024*1024)))
	if err != nil {
		return nil, fmt.Errorf("dial error on %v -> %w", "pstore.pstore:8080", err)
	}
	return &pClient{pClient: pb.NewPStoreServiceClient(conn)}, nil
}

func (c *pClient) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	return c.pClient.Read(ctx, req)
}

func (c *pClient) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	return c.pClient.Write(ctx, req)
}

func (c *pClient) GetKeys(ctx context.Context, req *pb.GetKeysRequest) (*pb.GetKeysResponse, error) {
	return c.pClient.GetKeys(ctx, req)
}

func (c *pClient) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	return c.pClient.Delete(ctx, req)
}

func (c *pClient) Count(ctx context.Context, req *pb.CountRequest) (*pb.CountResponse, error) {
	return c.pClient.Count(ctx, req)
}
