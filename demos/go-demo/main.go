package main

import (
	"fmt"
	"time"

	"consolehub/libraries/go/consolehub"
)

func main() {
	fmt.Println("Starting ConsoleHub Demo Agent...")

	client, err := consolehub.New(
		consolehub.WithEndpoint("ws://localhost:3787/api/v1/rpc/ws"),
		consolehub.WithTenant("acme"),
		consolehub.WithApp("demo-replicator"),
		consolehub.WithToken("demo-token"),
	)
	if err != nil {
		client, _ = consolehub.New(consolehub.WithDisabled(true))
	}
	defer client.Close()

	client.Println("Demo Agent registered and active")

	for i := 1; i <= 5; i++ {
		client.Printf("Downloading file artifact part %d...\n", i)
		client.Infof("Progress update: %d%%", i*20)
		time.Sleep(500 * time.Millisecond)
	}

	client.Println("Demo Agent execution finished cleanly.")
}
