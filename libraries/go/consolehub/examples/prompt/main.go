package main

import (
	"consolehub/libraries/go/consolehub"
)

func main() {
	defer consolehub.Close()

	consolehub.Println("Interactive Console Prompts Demo")

	// Text Prompt
	channel := consolehub.Prompt("Enter target channel username", "@mychannel")
	consolehub.Printf("Selected channel: %s\n", channel)

	// Choice Prompt
	env := consolehub.Choice("Select Deployment Environment", []string{"Development", "Staging", "Production"}, "Development")
	consolehub.Printf("Selected environment: %s\n", env)

	// Confirmation Prompt
	if consolehub.Confirm("Proceed with deployment?", true) {
		consolehub.Println("Deployment started.")
	} else {
		consolehub.Println("Deployment cancelled.")
	}
}
