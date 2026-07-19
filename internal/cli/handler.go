package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/caramelang/caramel/internal/composer"
	"github.com/caramelang/caramel/internal/parser/analyzer"
	"github.com/caramelang/caramel/pkg/branding"
	"github.com/caramelang/caramel/pkg/clifmt"
	"github.com/rp1s/digreyt/translate"
)

type Command func(args []string) error

var CurrentLanguage = translate.DefaultLanguage

var commands = map[string]Command{
	"build":   Build,
	"help":    Help,
	"install": Install,
}

func HandlerCmd() error {
	args, showHelp, err := parseGlobalArgs(os.Args[1:])
	if err != nil {
		return err
	}
	if showHelp || len(args) == 0 {
		return Help(nil)
	}

	cmd, ok := commands[args[0]]
	if !ok {
		return Help(nil)
	}

	return cmd(args[1:])
}

func SetLanguage(language string) {
	language = strings.ToLower(strings.TrimSpace(language))
	if language == "" {
		language = translate.DefaultLanguage
	}
	CurrentLanguage = language
	translate.SetLanguage(language)
}

func parseGlobalArgs(args []string) ([]string, bool, error) {
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			return out, true, nil
		case arg == "--lang" || arg == "-L":
			if i+1 >= len(args) {
				return nil, false, fmt.Errorf("usage: caramel --lang <lang> <command>")
			}
			i++
			SetLanguage(args[i])
		case strings.HasPrefix(arg, "--lang="):
			SetLanguage(strings.TrimPrefix(arg, "--lang="))
		default:
			out = append(out, arg)
		}
	}
	return out, false, nil
}

func Help(args []string) error {
	clifmt.Print(helpDocument(), CurrentLanguage)
	return nil
}

func Build(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: caramel build <path> [name]")
	}
	if len(args) > 2 {
		return fmt.Errorf("usage: caramel build <path> [name]")
	}

	project, err := composer.Load(args[0], projectNameArg(args))
	if err != nil {
		return err
	}
	result, err := analyzer.New().Project(project)
	if err != nil {
		return err
	}
	if result.Diagnostics.HasFatalErrors() {
		return result.Diagnostics
	}
	return nil
}

func Install(args []string) error {
	return fmt.Errorf("install command is not implemented yet")
}

func helpDocument() clifmt.Document {
	return clifmt.Document{
		Art:     Art(branding.ColorArt),
		ShowArt: true,
		Usage:   clifmt.T("caramel [options] <command> [arguments]"),
		Sections: []clifmt.Section{
			{
				Title: clifmt.T("Commands"),
				Rows: []clifmt.Row{
					{
						Label:       "build <path> [name]",
						Description: clifmt.T("Build and analyze a source file."),
						Children: []clifmt.Row{
							{Label: "<path>", Description: clifmt.T("File path relative to the current directory.")},
							{Label: "[name]", Description: clifmt.T("Project name; defaults to the file name without extension.")},
						},
					},
					{
						Label:       "install",
						Description: clifmt.T("Install or update packages and tools."),
						Children: []clifmt.Row{
							{Label: "--global", Description: clifmt.T("Use the global installation scope.")},
							{Label: "--update", Description: clifmt.T("Update an existing installation.")},
						},
					},
					{
						Label:       "help",
						Description: clifmt.T("Show this help message."),
					},
				},
			},
			{
				Title: clifmt.T("Global Options"),
				Rows: []clifmt.Row{
					{Label: "-h, --help", Description: clifmt.T("Show help and exit.")},
					{Label: "--lang <lang>, -L <lang>", Description: clifmt.T("Set output language, for example eng or ru.")},
				},
			},
			{
				Title: clifmt.T("Examples"),
				Rows: []clifmt.Row{
					{Label: "caramel build examples/models/model.cm", Description: clifmt.T("Build using the file name as project name.")},
					{Label: "caramel build src/user.cm user", Description: clifmt.T("Build with an explicit project name.")},
					{Label: "caramel --lang ru help", Description: clifmt.T("Show help in Russian.")},
				},
			},
		},
	}
}

func projectNameArg(args []string) string {
	if len(args) < 2 {
		return ""
	}
	return args[1]
}
