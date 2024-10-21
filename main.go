package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	pb "github.com/brotherlogic/pstore/proto"

	ghbclient "github.com/brotherlogic/githubridge/client"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	port        = flag.Int("port", 8080, "The server port.")
	metricsPort = flag.Int("metrics_port", 8081, "Metrics port")
)

var ()

type Server struct {
	gclient ghbclient.GithubridgeClient

	clients []pb.PStoreServiceClient
}

type rstore interface {
	Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error)
	Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error)
	GetKeys(ctx context.Context, req *pb.GetKeysRequest) (*pb.GetKeysResponse, error)
	Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error)
	Count(ctx context.Context, req *pb.CountRequest) (*pb.CountResponse, error)
}

func (s *Server) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	var reads []*pb.ReadResponse
	for _, c := range s.clients {
		resp, err := c.Read(ctx, req)
		if err != nil {
			log.Printf("Error on read: %v", err)
		}
		reads = append(reads, resp)
	}

	if len(reads) == 0 {
		return nil, status.Errorf(codes.Internal, "Unable to process %v", req)
	}

	return reads[0], nil
}

func (s *Server) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	var writes []*pb.WriteResponse
	for _, c := range s.clients {
		resp, err := c.Write(ctx, req)
		if err != nil {
			log.Printf("Error on read: %v", err)
		}
		writes = append(writes, resp)
	}

	if len(writes) == 0 {
		return nil, status.Errorf(codes.Internal, "Unable to process %v", req)
	}

	return writes[0], nil
}

func (s *Server) GetKeys(ctx context.Context, req *pb.GetKeysRequest) (*pb.GetKeysResponse, error) {
	var keys []*pb.GetKeysResponse
	for _, c := range s.clients {
		resp, err := c.GetKeys(ctx, req)
		if err != nil {
			log.Printf("Error on read: %v", err)
		}
		keys = append(keys, resp)
	}

	if len(keys) == 0 {
		return nil, status.Errorf(codes.Internal, "Unable to process %v", req)
	}

	return keys[0], nil
}

func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	var deletes []*pb.DeleteResponse
	for _, c := range s.clients {
		resp, err := c.Delete(ctx, req)
		if err != nil {
			log.Printf("Error on read: %v", err)
		}
		deletes = append(deletes, resp)
	}

	if len(deletes) == 0 {
		return nil, status.Errorf(codes.Internal, "Unable to process %v", req)
	}

	return deletes[0], nil
}

func (s *Server) Count(ctx context.Context, req *pb.CountRequest) (*pb.CountResponse, error) {
	var counts []*pb.CountResponse
	for _, c := range s.clients {
		resp, err := c.Count(ctx, req)
		if err != nil {
			log.Printf("Error on read: %v", err)
		}
		counts = append(counts, resp)
	}

	if len(counts) == 0 {
		return nil, status.Errorf(codes.Internal, "Unable to process %v", req)
	}

	return counts[0], nil
}

func main() {
	flag.Parse()

	s := &Server{}

	// Register the rstore client here

	client, err := ghbclient.GetClientInternal()
	if err != nil {
		log.Fatalf("Unable to reach GHB")
	}
	s.gclient = client

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("pstore failed to listen on the serving port %v: %v", *port, err)
	}
	size := 1024 * 1024 * 1000
	gs := grpc.NewServer(
		grpc.MaxSendMsgSize(size),
		grpc.MaxRecvMsgSize(size),
	)
	pb.RegisterPStoreServiceServer(gs, s)
	log.Printf("pstore is listening on %v", lis.Addr())

	// Setup prometheus export
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(fmt.Sprintf(":%v", *metricsPort), nil)
	}()

	if err := gs.Serve(lis); err != nil {
		log.Fatalf("pstore failed to serve: %v", err)
	}
}
