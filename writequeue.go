package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/protobuf/types/known/anypb"

	pb "github.com/brotherlogic/pstore/proto"
)

type WriteElement struct {
	key   string
	value []byte
	cname string
}

func (s *Server) runWriteQueue() {
	for {
		elem := <-s.wq
		s.runElem(elem)
	}
}

func (s *Server) runElem(we *WriteElement) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	for _, c := range s.clients {
		if c.Name() == we.cname {
			_, err := c.Write(ctx, &pb.WriteRequest{
				Key:   we.key,
				Value: &anypb.Any{Value: we.value},
			})
			log.Printf("Side Write (%v, %v) -> %v", we.cname, we.key, err)
		}
	}
}
