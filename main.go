package main

import (
	"fmt"
	"os"

	"mogura/internal/app"
)

func main() {
	cfg, err := app.ParseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := app.Run(cfg, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
