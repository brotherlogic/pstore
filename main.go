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
	pb "github.com/brotherlogic/pstore/proto"
	rstore_client "github.com/brotherlogic/rstore/client"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
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
	dCountDiffs = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pstore_delete_diffs",
	})

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

func (s *Server) runRead(ctx context.Context, client pstore, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	t := time.Now()
	resp, err := client.Read(ctx, req)
	rCount.With(prometheus.Labels{"client": client.Name(), "code": fmt.Sprintf("%v", status.Code(err))}).Inc()
	if err == nil {
		rCountTime.With(prometheus.Labels{"client": client.Name()}).Observe(float64(time.Since(t).Milliseconds()))
	}
	return resp, err
}

func (s *Server) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	mResp, err := s.runRead(ctx, s.clients[0], req)

	if err == nil {
		for _, c := range s.clients[1:] {
			go func() {
				resp, err := s.runRead(ctx, c, req)
				if err == nil {
					if len(resp.GetValue().GetValue()) != len(mResp.GetValue().GetValue()) {
						rCountDiffs.Inc()
					}
				}
			}()
		}
	}

	return mResp, err
}

func (s *Server) runWrite(ctx context.Context, client pstore, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	t := time.Now()
	resp, err := client.Write(ctx, req)
	wCount.With(prometheus.Labels{"client": client.Name(), "code": fmt.Sprintf("%v", status.Code(err))}).Inc()
	if err == nil {
		wCountTime.With(prometheus.Labels{"client": client.Name()}).Observe(float64(time.Since(t).Milliseconds()))
	}
	return resp, err
}

func (s *Server) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	mresp, err := s.runWrite(ctx, s.clients[0], req)
	if err == nil {
		for _, c := range s.clients[1:] {
			go func() {
				s.runWrite(ctx, c, req)
			}()
		}
	}
	return mresp, err
}

func (s *Server) runGetKeys(ctx context.Context, client pstore, req *pb.GetKeysRequest) (*pb.GetKeysResponse, error) {
	t := time.Now()
	resp, err := client.GetKeys(ctx, req)
	gkCount.With(prometheus.Labels{"client": client.Name(), "code": fmt.Sprintf("%v", status.Code(err))}).Inc()
	if err == nil {
		gkCountTime.With(prometheus.Labels{"client": client.Name()}).Observe(float64(time.Since(t).Milliseconds()))
	}
	return resp, err
}

func (s *Server) GetKeys(ctx context.Context, req *pb.GetKeysRequest) (*pb.GetKeysResponse, error) {
	mresp, err := s.runGetKeys(ctx, s.clients[0], req)
	if err == nil {
		for _, c := range s.clients[1:] {
			go func() {
				resp, err := s.runGetKeys(ctx, c, req)
				if err == nil {
					if len(resp.GetKeys()) != len(mresp.GetKeys()) {
						gkCountDiffs.Inc()
					}
				}
			}()
		}
	}
	return mresp, err
}

func (s *Server) runDelete(ctx context.Context, client pstore, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	t := time.Now()
	resp, err := client.Delete(ctx, req)
	dCount.With(prometheus.Labels{"client": client.Name(), "code": fmt.Sprintf("%v", status.Code(err))}).Inc()
	if err == nil {
		dCountTime.With(prometheus.Labels{"client": client.Name()}).Observe(float64(time.Since(t).Milliseconds()))
	}
	return resp, err
}

func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	mresp, err := s.runDelete(ctx, s.clients[0], req)
	for _, c := range s.clients[1:] {
		go func() {
			s.runDelete(ctx, c, req)
		}()
	}
	return mresp, err
}

func (s *Server) runCount(ctx context.Context, client pstore, req *pb.CountRequest) (*pb.CountResponse, error) {
	t := time.Now()
	resp, err := client.Count(ctx, req)
	cCount.With(prometheus.Labels{"client": client.Name(), "code": fmt.Sprintf("%v", status.Code(err))}).Inc()
	if err == nil {
		cCountTime.With(prometheus.Labels{"client": client.Name()}).Observe(float64(time.Since(t).Milliseconds()))
	}
	return resp, err
}

func (s *Server) Count(ctx context.Context, req *pb.CountRequest) (*pb.CountResponse, error) {
	mresp, err := s.runCount(ctx, s.clients[0], req)

	if err == nil {
		for _, c := range s.clients {
			go func() {
				resp, err := s.runCount(ctx, c, req)
				if err == nil {
					if resp.GetCount() != mresp.GetCount() {
						cCountDiffs.Inc()
					}
				}
			}()
		}
	}

	return mresp, err
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

	pgc, err := getPGStore()
	if err != nil {
		log.Fatalf("Cannot dial pgstore client")
	}
	s.clients = append(s.clients, pgc)
	/*msc, err := mstore_client.GetClient()
	if err != nil {
		log.Fatalf("Unable to get mstore client")
	}
	s.clients = append(s.clients, &mstore_wrapper{mc: msc})*/

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
