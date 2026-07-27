package main

import (
	"time"

	"consolehub/libraries/go/consolehub"
)

func main() {
	// Initialize ConsoleHub telemetry with explicit tenant and app metadata
	client, err := consolehub.New(
		consolehub.WithTenant("telegram-ops"),
		consolehub.WithApp("telegram-replicator"),
		consolehub.WithEndpoint("ws://localhost:3787/api/v1/rpc/ws"),
	)
	if err != nil {
		// Degrades gracefully
		client, _ = consolehub.New(consolehub.WithDisabled(true))
	}
	defer client.Close()

	client.Println("=== Telegram Channel Message Replicator ===")
	client.Infof("Connecting to Telegram MTProto session...")

	// 1. Progress Bar for File Downloading
	p := client.Progress("Downloading channel media files", 50)
	for i := 0; i <= 50; i += 10 {
		p.Set(int64(i))
		time.Sleep(40 * time.Millisecond)
	}
	p.Done()

	// 2. Structured Ingestion Logs
	client.Infof("Replicated 450 messages from channel @crypto_alerts")
	client.Warnf("Rate limit threshold 80%% reached on peer connection")

	client.Println("=== Replication Session Finished Successfully ===")
}
