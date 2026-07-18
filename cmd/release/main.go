package main

import (
	"context"
	"fmt"
	"os"

	"github.com/CandyCrafts/candy/internal/cli"
	"github.com/CandyCrafts/candy/internal/database"
	"github.com/CandyCrafts/candy/pkg/branding"
	"github.com/CandyCrafts/candy/pkg/clifmt"
)

func main() {
	conn, err := database.OpenDatabase(branding.DatabaseFileName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer conn.Close()

	cdb := database.NewCacheDatabase(conn)
	err = cdb.Init(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	clifmt.SetDefaultCacheStore(cdb.CLIText())

	err = cli.HandlerCmd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
