package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/samirgadkari/documents/pkg/config"
	"github.com/samirgadkari/documents/pkg/data"
	"github.com/samirgadkari/sidecar/pkg/client"
	"github.com/samirgadkari/sidecar/pkg/utils"
	pb "github.com/samirgadkari/sidecar/protos/v1/messages"
)

const (
	allTopicsRecvChanSize = 32
)

func main() {

	config.Load()

	// Setup database
	db, err := data.DBConnect()
	if err != nil {
		return
	}

	tableName := "documents"
	err = db.CreateDocumentsTable()
	if err != nil {
		return
	}

	sidecar, err := client.InitSidecar(tableName, nil)
	if err != nil {
		fmt.Printf("Error initializing sidecar: %v\n", err)
		os.Exit(-1)
	}

	topic := "search.data.v1"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = sidecar.ProcessSubMsgs(ctx, topic,
		allTopicsRecvChanSize, func(m *pb.SubTopicResponse) {

			msg := fmt.Sprintf("Received from sidecar:\n\t%s", m.String())
			fmt.Printf("%s\n", msg)

			// db.StoreData(m.Header, msg, tableName)
		})
	if err != nil {
		fmt.Printf("Error processing subscription messages:\n\ttopic: %s\n\terr: %v\n",
			topic, err)
	}

	/* This is an example of how to publish a message. It is a log message because for now
	* it is the only type that is received (by this same persistLogs service).

	sidecar.Logger.Log("Persist sending log message test: %s\n", "search.log.v1")
	time.Sleep(3 * time.Second)

	var retryNum uint32 = 1
	retryDelayDuration, err := time.ParseDuration("200ms")
	if err != nil {
		fmt.Printf("Error creating Golang time duration.\nerr: %v\n", err)
		os.Exit(-1)
	}
	retryDelay := durationpb.New(retryDelayDuration)

	err = sidecar.Pub(ctx, "search.data.v1", []byte("test pub message"),
		&pb.RetryBehavior{
			RetryNum:   &retryNum,
			RetryDelay: retryDelay,
		},
	)
	if err != nil {
		fmt.Printf("Error publishing message.\n\terr: %v\n", err)
	}

	*/

	fmt.Println("Press the Enter key to stop")
	fmt.Scanln()
	fmt.Println("User pressed Enter key")

	// Signal that we want the process subscription goroutines to end.
	// This cancellation causes the goroutines to unsubscribe from the topic
	// before they end themselves.
	cancel()

	sleepDur, _ := time.ParseDuration("3s")
	fmt.Printf("Sleeping for %s seconds\n", sleepDur)
	time.Sleep(sleepDur)

	utils.ListGoroutinesRunning()

	select {} // This will wait forever
}