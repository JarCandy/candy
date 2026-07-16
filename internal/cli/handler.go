package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/CandyCrafts/candy/internal/analyzer"
	"github.com/CandyCrafts/candy/internal/composer"
	"github.com/CandyCrafts/candy/pkg/clifmt"
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
				return nil, false, fmt.Errorf("usage: candy --lang <lang> <command>")
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
		return fmt.Errorf("usage: candy build <path> [name]")
	}
	if len(args) > 2 {
		return fmt.Errorf("usage: candy build <path> [name]")
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
		Art:         candyArt,
		ArtGradient: CandyGradientArt,
		Usage: clifmt.T(
			"candy [options] <command> [arguments]",
			clifmt.Lang("ru", "candy [опции] <команда> [аргументы]"),
		),
		Sections: []clifmt.Section{
			{
				Title: clifmt.T("Commands", clifmt.Lang("ru", "Команды")),
				Rows: []clifmt.Row{
					{
						Label: "build <path> [name]",
						Description: clifmt.T(
							"Build and analyze a source file.",
							clifmt.Lang("ru", "Собрать и проанализировать файл."),
						),
						Children: []clifmt.Row{
							{Label: "<path>", Description: clifmt.T("File path relative to the current directory.", clifmt.Lang("ru", "Путь к файлу относительно текущей директории."))},
							{Label: "[name]", Description: clifmt.T("Project name; defaults to the file name without extension.", clifmt.Lang("ru", "Имя проекта; по умолчанию имя файла без расширения."))},
						},
					},
					{
						Label: "install",
						Description: clifmt.T(
							"Install or update packages and tools.",
							clifmt.Lang("ru", "Установить или обновить пакеты и инструменты."),
						),
						Children: []clifmt.Row{
							{Label: "--global", Description: clifmt.T("Use the global installation scope.", clifmt.Lang("ru", "Использовать глобальную область установки."))},
							{Label: "--update", Description: clifmt.T("Update an existing installation.", clifmt.Lang("ru", "Обновить существующую установку."))},
						},
					},
					{
						Label: "help",
						Description: clifmt.T(
							"Show this help message.",
							clifmt.Lang("ru", "Показать это сообщение справки."),
						),
					},
				},
			},
			{
				Title: clifmt.T("Global Options", clifmt.Lang("ru", "Глобальные опции")),
				Rows: []clifmt.Row{
					{Label: "-h, --help", Description: clifmt.T("Show help and exit.", clifmt.Lang("ru", "Показать справку и выйти."))},
					{Label: "--lang <lang>, -L <lang>", Description: clifmt.T("Set output language, for example eng or ru.", clifmt.Lang("ru", "Задать язык вывода, например eng или ru."))},
				},
			},
			{
				Title: clifmt.T("Examples", clifmt.Lang("ru", "Примеры")),
				Rows: []clifmt.Row{
					{Label: "candy build examples/models/model.cm", Description: clifmt.T("Build using the file name as project name.", clifmt.Lang("ru", "Собрать, используя имя файла как имя проекта."))},
					{Label: "candy build src/user.cm user", Description: clifmt.T("Build with an explicit project name.", clifmt.Lang("ru", "Собрать с явным именем проекта."))},
					{Label: "candy --lang ru help", Description: clifmt.T("Show help in Russian.", clifmt.Lang("ru", "Показать справку на русском."))},
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
