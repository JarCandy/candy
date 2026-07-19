package main

import (
	"context"
	stderrors "errors"
	"fmt"
	"os"

	"github.com/caramelang/caramel/internal/cli"
	"github.com/caramelang/caramel/internal/database"
	"github.com/caramelang/caramel/pkg/branding"
	"github.com/caramelang/caramel/pkg/clifmt"
)

func main() {
	if err := run(); err != nil {
		if printErr := cli.PrintError(os.Stderr, err); printErr != nil {
			fmt.Fprintln(os.Stderr, printErr)
		}
		os.Exit(1)
	}
}

func run() (resultErr error) {
	conn, err := database.OpenDatabase(branding.DatabaseFileName)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			resultErr = stderrors.Join(resultErr, fmt.Errorf("close database: %w", closeErr))
		}
	}()

	cdb := database.NewCacheDatabase(conn)
	if err := cdb.Init(context.Background()); err != nil {
		return err
	}

	clifmt.SetDefaultCacheStore(cdb.CLIText())
	return cli.HandlerCmd()
}
