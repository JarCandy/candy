package cli

import (
	"os"
	"strings"

	"github.com/CandyCrafts/candy/internal/project"
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

	sln(sb, cls.Gradient(candyArt, candyGradientArt))
	sln(sb, cls.Gradient("Usage: candy <command> [options]\n", candyGradient))

	sln(sb, cls.Gradient("Commands:", candyGradient))
	s(sb, "  build")
	sln(sb, cls.Apply("   Build the project", colorista.Rgb(colorista.RGB{R: 217, G: 217, B: 217})))
	s(sb, "  help")
	sln(sb, cls.Apply("    Show this help message\n", colorista.Rgb(colorista.RGB{R: 217, G: 217, B: 217})))
	sln(sb, cls.Gradient("Options:", candyGradient))
	s(sb, "  -h, --help")
	sln(sb, cls.Apply("   Show this help message", colorista.Rgb(colorista.RGB{R: 217, G: 217, B: 217})))

	println(sb.String())
	return nil
}

func Build() error {
	project, err := project.Load()
	if err != nil {
		return err
	}
	_ = project
	return nil
}
