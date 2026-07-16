package main

import (
	"fmt"
	"os"

	"github.com/CandyCrafts/candy/internal/cli"
)

func main() {
	err := cli.HandlerCmd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
