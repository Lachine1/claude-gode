package main

import (
	"fmt"
	"os"

	"github.com/Lachine1/claude-gode/internal/cli"
	"github.com/Lachine1/claude-gode/pkg/types"
)

var (
	version   = types.Version
	buildTime = types.BuildTime
)

func main() {
	if err := cli.Run(os.Args[1:], version, buildTime); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
