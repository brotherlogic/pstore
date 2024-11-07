package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	ghbclient "github.com/brotherlogic/githubridge/client"
	mstore_client "github.com/brotherlogic/mstore/client"
	pb "github.com/brotherlogic/pstore/proto"
	rstore_client "github.com/brotherlogic/rstore/client"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	port        = flag.Int("port", 8080, "The server port.")
	metricsPort = flag.Int("metrics_port", 8081, "Metrics port")
)

var (
	wCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pstore_wcount",
	}, []string{"client", "code"})
	wCountTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "pstore_wcount_latency",
	}, []string{"client"})

	dCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pstore_dcount",
	}, []string{"client", "code"})
	dCountTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "pstore_dcount_latency",
	}, []string{"client"})

	rCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pstore_rcount",
	}, []string{"client", "code"})
	rCountTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "pstore_rcount_latency",
	}, []string{"client"})
	rCountDiffs = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pstore_rcount_diffs",
	})

	gkCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pstore_gkcount",
	}, []string{"client", "code"})
	gkCountTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "pstore_gkcount_latency",
	}, []string{"client"})
	gkCountDiffs = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pstore_gkcount_diffs",
	})

	cCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pstore_ccount",
	}, []string{"client", "code"})
	cCountTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "pstore_ccount_latency",
	}, []string{"client"})
	cCountDiffs = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pstore_ccount_diffs",
	})

	cSplit = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pstore_split",
	})
)

type Server struct {
	gclient ghbclient.GithubridgeClient

	clients []pstore
}

type pstore interface {
	Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error)
	Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error)
	GetKeys(ctx context.Context, req *pb.GetKeysRequest) (*pb.GetKeysResponse, error)
	Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error)
	Count(ctx context.Context, req *pb.CountRequest) (*pb.CountResponse, error)
	Name() string
}

func (s *Server) split(ctx context.Context, arlen int) int {
	splitVal := float64(0.0)
	cSplit.Set(splitVal)

	if arlen == 0 {
		return 0
	}

	if rand.Float64() > splitVal {
		return 0
	}
	return 1
}

func (s *Server) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	var reads []*pb.ReadResponse
	var errors []error
	for _, c := range s.clients {
		t := time.Now()
		resp, err := c.Read(ctx, req)
		rCount.With(prometheus.Labels{"client": c.Name(), "code": fmt.Sprintf("%v", status.Code(err))}).Inc()

		if err != nil {
			log.Printf("Error on read: %v", err)
		} else {
			rCountTime.With(prometheus.Labels{"client": c.Name()}).Observe(float64(time.Since(t).Milliseconds()))
		}

		reads = append(reads, resp)
		errors = append(errors, err)
	}

	for i, val := range reads[1:] {
		if errors[i+1] == nil {
			if len(val.GetValue().GetValue()) != len(reads[0].GetValue().GetValue()) {
				rCountDiffs.Inc()
				log.Printf("Diff: %v", req.GetKey())
				break
			}

			for i := range val.GetValue().GetValue() {
				if val.GetValue().GetValue()[i] != reads[0].GetValue().GetValue()[i] {
					rCountDiffs.Inc()
					log.Printf("Diff: %v", req.GetKey())
					break
				}
			}
		}
	}

	return reads[0], errors[0]
}

func (s *Server) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	var writes []*pb.WriteResponse
	var errors []error
	for _, c := range s.clients {
		t := time.Now()
		resp, err := c.Write(ctx, req)
		wCount.With(prometheus.Labels{"client": c.Name(), "code": fmt.Sprintf("%v", status.Code(err))}).Inc()
		if err != nil {
			log.Printf("Error on write: %v", err)
		} else {
			log.Printf("Written in %v (%v)", time.Since(t), c.Name())
			wCountTime.With(prometheus.Labels{"client": c.Name()}).Observe(float64(time.Since(t).Milliseconds()))
		}
		writes = append(writes, resp)
		errors = append(errors, err)
	}

	if len(writes) == 0 {
		return nil, status.Errorf(codes.Internal, "Unable to process %v", req)
	}

	i := s.split(ctx, len(writes))
	return writes[i], errors[i]
}

func (s *Server) GetKeys(ctx context.Context, req *pb.GetKeysRequest) (*pb.GetKeysResponse, error) {
	var keys []*pb.GetKeysResponse
	var errors []error
	for _, c := range s.clients {
		t := time.Now()
		resp, err := c.GetKeys(ctx, req)
		gkCount.With(prometheus.Labels{"client": c.Name(), "code": fmt.Sprintf("%v", status.Code(err))}).Inc()

		if err != nil {
			log.Printf("Error on read: %v", err)
		} else {
			gkCountTime.With(prometheus.Labels{"client": c.Name()}).Observe(float64(time.Since(t).Milliseconds()))
		}
		keys = append(keys, resp)
		errors = append(errors, err)
	}

	if len(keys) == 0 {
		return nil, status.Errorf(codes.Internal, "Unable to process %v", req)
	}

	for i, val := range keys[1:] {
		if errors[i+1] == nil {
			if len(val.GetKeys()) != len(keys[0].GetKeys()) {
				rCountDiffs.Inc()
			}

			for i := range val.GetKeys() {
				if val.GetKeys()[i] != keys[0].GetKeys()[i] {
					gkCountDiffs.Inc()
					break
				}
			}
		}
	}
	i := s.split(ctx, len(keys))
	return keys[i], errors[i]
}

func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	var deletes []*pb.DeleteResponse
	var errors []error
	for _, c := range s.clients {
		t := time.Now()
		resp, err := c.Delete(ctx, req)
		dCount.With(prometheus.Labels{"client": c.Name(), "code": fmt.Sprintf("%v", status.Code(err))}).Inc()

		if err != nil {
			log.Printf("Error on read: %v", err)
		} else {
			dCountTime.With(prometheus.Labels{"client": c.Name()}).Observe(float64(time.Since(t).Milliseconds()))
		}
		deletes = append(deletes, resp)
		errors = append(errors, err)
	}

	if len(deletes) == 0 {
		return nil, status.Errorf(codes.Internal, "Unable to process %v", req)
	}

	i := s.split(ctx, len(deletes))
	return deletes[i], errors[i]
}

func (s *Server) Count(ctx context.Context, req *pb.CountRequest) (*pb.CountResponse, error) {
	var counts []*pb.CountResponse
	var errors []error
	for _, c := range s.clients {
		t := time.Now()
		resp, err := c.Count(ctx, req)
		cCount.With(prometheus.Labels{"client": c.Name(), "code": fmt.Sprintf("%v", status.Code(err))}).Inc()
		if err != nil {
			log.Printf("Error on read: %v", err)
		} else {
			cCountTime.With(prometheus.Labels{"client": c.Name()}).Observe(float64(time.Since(t).Milliseconds()))
		}
		counts = append(counts, resp)
		errors = append(errors, err)
	}

	if len(counts) == 0 {
		return nil, status.Errorf(codes.Internal, "Unable to process %v", req)
	}

	val := counts[0].GetCount()
	for i, c := range counts[1:] {
		if c.GetCount() != val && errors[i+1] == nil {
			cCountDiffs.Inc()
		}
	}

	i := s.split(ctx, len(counts))
	return counts[i], errors[i]
}

func main() {
	flag.Parse()

	s := &Server{}

	// Register the rstore client here
	rsc, err := rstore_client.GetClient()
	if err != nil {
		log.Fatalf("Unable to reach rstore client")
	}
	s.clients = append(s.clients, &rstore_wrapper{rc: rsc})

	msc, err := mstore_client.GetClient()
	if err != nil {
		log.Fatalf("Unable to get mstore client")
	}
	s.clients = append(s.clients, &mstore_wrapper{mc: msc})

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
