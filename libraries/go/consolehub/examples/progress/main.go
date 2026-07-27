package main

import (
	"time"

	"consolehub/libraries/go/consolehub"
)

func main() {
	defer consolehub.Close()

	consolehub.Println("Starting Progress Tracking Demo...")

	p := consolehub.Progress("Downloading dataset archive", 100)
	for i := 0; i <= 100; i += 10 {
		p.Set(int64(i))
		time.Sleep(30 * time.Millisecond)
	}
	p.Done()

	up := consolehub.UploadProgress("Uploading processed artifacts", 10485760)
	for bytesSent := int64(0); bytesSent <= 10485760; bytesSent += 1048576 {
		up.Set(bytesSent)
		time.Sleep(30 * time.Millisecond)
	}
	up.Finish()

	consolehub.Println("Progress tracking completed.")
}
