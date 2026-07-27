package main

import (
	"log"

	"github.com/aognio/consolehub/libraries/go/consolehub"
)

func main() {
	defer consolehub.Close()

	// Direct ConsoleHub structured log methods
	consolehub.Debug("Configuring worker pool")
	consolehub.Infof("Worker pool size: %d", 4)
	consolehub.Warn("High memory load detected")
	consolehub.Errorf("Network timeout to upstream node")

	// Standard Go log package redirected through ConsoleHub
	stdLogger := log.New(consolehub.Stdout(), "[STD-LOG] ", log.LstdFlags)
	stdLogger.Println("Routed through Go stdlib log package into ConsoleHub")
}
