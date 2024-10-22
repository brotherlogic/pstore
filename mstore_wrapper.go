package main

import (
	"context"

	mstore_client "github.com/brotherlogic/mstore/client"

	mspb "github.com/brotherlogic/mstore/proto"
	pb "github.com/brotherlogic/pstore/proto"
)

type mstore_wrapper struct {
	mc mstore_client.MStoreClient
}

func (r *mstore_wrapper) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	resp, err := r.mc.Write(ctx, &mspb.WriteRequest{
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

func (r *mstore_wrapper) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	resp, err := r.mc.Read(ctx, &mspb.ReadRequest{
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

func (r *mstore_wrapper) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	_, err := r.mc.Delete(ctx, &mspb.DeleteRequest{
		Key: req.GetKey(),
	})
	if err != nil {
		return &pb.DeleteResponse{}, err
	}

	return &pb.DeleteResponse{}, nil
}

func (r *mstore_wrapper) GetKeys(ctx context.Context, req *pb.GetKeysRequest) (*pb.GetKeysResponse, error) {
	resp, err := r.mc.GetKeys(ctx, &mspb.GetKeysRequest{
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

func (r *mstore_wrapper) Count(ctx context.Context, req *pb.CountRequest) (*pb.CountResponse, error) {
	resp, err := r.rm.Count(ctx, &mspb.CountRequest{
		Counter: req.GetCounter(),
	})
	if err != nil {
		return &pb.CountResponse{}, err
	}

	return &pb.CountResponse{
		Count: resp.GetCount(),
	}, nil
}
