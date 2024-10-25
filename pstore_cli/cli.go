package main

import (
	"log"
	"os"
	"time"

	pbps "github.com/brotherlogic/pstore/proto"

	"github.com/brotherlogic/goserver/utils"
)

func main() {
	ctx, cancel := utils.ManualContext("pstore-cli", time.Hour)
	defer cancel()

	conn, err := utils.LFDial(os.Args[1])
	if err != nil {
		log.Fatalf("Bad dial: %v", err)
	}

	client := pbps.NewPStoreServiceClient(conn)

	result, err := client.Read(ctx, &pbps.ReadRequest{Key: "sync/test"})
	if err != nil {
		log.Printf("Error: %v", err)
	}
	log.Printf("Result: %v", result)
}
