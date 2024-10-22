package main

import (
	"context"

	rstore_client "github.com/brotherlogic/rstore/client"

	pb "github.com/brotherlogic/pstore/proto"
	rspb "github.com/brotherlogic/rstore/proto"
)

type rstore_wrapper struct {
	rc rstore_client.RStoreClient
}

func (r *rstore_wrapper) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	resp, err := r.rc.Write(ctx, &rspb.WriteRequest{
		Key:   req.GetKey(),
		Value: req.GetValue(),
	})
	if err != nil {
		return &pb.WriteResponse{}, err
	}

	return &pb.WriteResponse{
		Timestamp: resp.GetTimestamp(),
	}, nil
}

func (r *rstore_wrapper) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	resp, err := r.rc.Read(ctx, &rspb.ReadRequest{
		Key: req.GetKey(),
	})
	if err != nil {
		return &pb.ReadResponse{}, err
	}

	return &pb.ReadResponse{
		Timestamp: resp.GetTimestamp(),
		Value:     resp.GetValue(),
	}, nil
}

func (r *rstore_wrapper) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	_, err := r.rc.Delete(ctx, &rspb.DeleteRequest{
		Key: req.GetKey(),
	})
	if err != nil {
		return &pb.DeleteResponse{}, err
	}

	return &pb.DeleteResponse{}, nil
}

func (r *rstore_wrapper) GetKeys(ctx context.Context, req *pb.GetKeysRequest) (*pb.GetKeysResponse, error) {
	resp, err := r.rc.GetKeys(ctx, &rspb.GetKeysRequest{
		AvoidSuffix: req.GetAvoidSuffix(),
		Prefix:      req.GetPrefix(),
		AllKeys:     req.GetAllKeys(),
	})
	if err != nil {
		return &pb.GetKeysResponse{}, err
	}

	return &pb.GetKeysResponse{
		Keys: resp.GetKeys(),
	}, nil
}

func (r *rstore_wrapper) Count(ctx context.Context, req *pb.CountRequest) (*pb.CountResponse, error) {
	resp, err := r.rc.Count(ctx, &rspb.CountRequest{
		Counter: req.GetCounter(),
	})
	if err != nil {
		return &pb.CountResponse{}, err
	}

	return &pb.CountResponse{
		Count: resp.GetCount(),
	}, nil
}
