package main

import (
	"fmt"
	"io"
	"os"

	"github.com/aognio/consolehub/libraries/go/consolehub"
)

func main() {
	defer consolehub.Close()

	// MultiWriter sending to both local terminal and ConsoleHub stream
	mw := io.MultiWriter(os.Stdout, consolehub.Stdout())

	fmt.Fprintln(mw, "This text outputs to BOTH terminal and ConsoleHub simultaneously.")
	fmt.Fprintf(mw, "Formatted multiwriter value: %d\n", 4096)
}
