package main

import (
	"fmt"
	"os"

	"github.com/CandyCrafts/candy/internal/cli"
)

func main() {
	if err := cli.HandlerCmd(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
