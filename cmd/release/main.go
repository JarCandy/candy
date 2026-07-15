package main

import "github.com/CandyCrafts/candy/internal/cli"

func main() {
	err := cli.HandlerCmd()
	if err != nil {
		return
	}

}
