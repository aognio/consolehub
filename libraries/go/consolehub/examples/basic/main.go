package main

import (
	"time"

	"github.com/aognio/consolehub/libraries/go/consolehub"
)

func main() {
	defer consolehub.Close()

	consolehub.Println("Starting Basic ConsoleHub Example App...")
	for i := 1; i <= 5; i++ {
		consolehub.Printf("Processing item %d of 5...\n", i)
		time.Sleep(50 * time.Millisecond)
	}
	consolehub.Println("Basic example finished successfully.")
}
