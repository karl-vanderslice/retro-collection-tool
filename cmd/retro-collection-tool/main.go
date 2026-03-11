package main

import (
	"fmt"
	"os"

	"github.com/karl-vanderslice/retro-collection-tool/internal/app"
)

func main() {
	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
