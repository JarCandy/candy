package main

import (
	"fmt"
	"os"

	"github.com/caramelang/caramel/internal/cli"
	"github.com/caramelang/caramel/internal/parser"
	"github.com/rp1s/lipa"
)

type ParserView struct {
	File        string
	AST         any
	Diagnostics any
	Source      string
}

func main() {
	if err := run(); err != nil {
		if printErr := cli.PrintError(os.Stderr, err); printErr != nil {
			fmt.Fprintln(os.Stderr, printErr)
		}
		os.Exit(1)
	}
}

func run() error {
	fileName := "examples/models/2-model.cm"
	if len(os.Args) > 1 {
		fileName = os.Args[1]
	}

	source, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	p := parser.New(source, fileName)
	ast, parseErr := p.Run()
	if parseErr != nil {
		return parseErr
	}
	diagnostics := append([]any(nil), diagnosticsSnapshot(p.Diagnostics.Errors)...)

	options := []lipa.Option{lipa.WithTitle("Caramel Parser Tree")}
	if outputPath := os.Getenv("LIPA_OUT"); outputPath != "" {
		options = append(options, lipa.WithOutputPath(outputPath))
	}
	if os.Getenv("LIPA_NO_OPEN") != "" {
		options = append(options, lipa.WithoutOpen())
	}

	return lipa.View(ParserView{
		File:        fileName,
		AST:         ast,
		Diagnostics: diagnostics,
		Source:      string(source),
	}, options...)
}

func diagnosticsSnapshot[T any](items []T) []any {
	values := make([]any, 0, len(items))
	for i := range items {
		values = append(values, items[i])
	}
	return values
}
