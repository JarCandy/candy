package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/CandyCrafts/candy/internal/analyzer"
	"github.com/CandyCrafts/candy/internal/composer"
	"github.com/rp1s/colorista"
)

type Command func() error

var commands = map[string]Command{
	"build": Build,
	"help":  Help,
}

func HandlerCmd() error {
	if len(os.Args) < 2 {
		return Help()
	}

	cmd, ok := commands[os.Args[1]]
	if !ok {
		return Help()
	}

	return cmd()
}

func s(sb *strings.Builder, text string) {
	sb.WriteString(text)
}

func sln(sb *strings.Builder, text string) {
	s(sb, text)
	s(sb, "\n")
}

func Help() error {
	sb := &strings.Builder{}
	cls := colorista.NewColorista(colorista.ThemeAuto)

	sln(sb, cls.Gradient(candyArt, CandyGradientArt))
	sln(sb, cls.Gradient("Usage: candy <command> [options]\n", candyGradient))

	sln(sb, cls.Gradient("Commands:", candyGradient))
	s(sb, "  build <path> [name]")
	sln(sb, cls.Apply("   Build a file from the current directory", colorista.Rgb(colorista.RGB{R: 217, G: 217, B: 217})))
	s(sb, "  help")
	sln(sb, cls.Apply("    Show this help message\n", colorista.Rgb(colorista.RGB{R: 217, G: 217, B: 217})))
	sln(sb, cls.Gradient("Options:", candyGradient))
	s(sb, "  -h, --help")
	sln(sb, cls.Apply("   Show this help message", colorista.Rgb(colorista.RGB{R: 217, G: 217, B: 217})))

	fmt.Print(sb.String())
	return nil
}

func Build() error {
	args := os.Args[2:]
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

func projectNameArg(args []string) string {
	if len(args) < 2 {
		return ""
	}
	return args[1]
}
