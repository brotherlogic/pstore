package main

import (
	"log"
	"os"
	"time"

	pbps "github.com/brotherlogic/pstore/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/brotherlogic/goserver/utils"
)

func main() {
	ctx, cancel := utils.ManualContext("pstore-cli", time.Hour)
	defer cancel()

	size := 1024 * 1024 * 2000
	conn, err := grpc.Dial(os.Args[1], grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(size)))
	if err != nil {
		log.Fatalf("Bad dial: %v", err)
	}

	client := pbps.NewPStoreServiceClient(conn)

	result, err := client.GetKeys(ctx, &pbps.GetKeysRequest{AllKeys: true})
	if err != nil {
		log.Printf("Error: %v", err)
	}
	log.Printf("Found %v keys", len(result.GetKeys()))
	/*for _, key := range result.GetKeys() {
		log.Printf("Key: %v", key)
	}*/
}
