package cli

import (
	"os"
	"strings"

	"github.com/rp1s/colorista"
)

type Command func() error

var commands = map[string]Command{
	// "build": Build,
	"help": Help,
}

func HandlerCmd() error {
	if len(os.Args) < 1 {
		cmd := commands[os.Args[1]]
		return cmd()

	}
	cmd := commands[os.Args[1]]
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

	candyGradient := []colorista.GradientPos{
		{Pos: 0.00, Color: colorista.RGB{R: 255, G: 80, B: 180}},  // pink
		{Pos: 0.20, Color: colorista.RGB{R: 255, G: 120, B: 220}}, // candy
		{Pos: 0.40, Color: colorista.RGB{R: 170, G: 80, B: 255}},  // purple
		{Pos: 0.60, Color: colorista.RGB{R: 80, G: 180, B: 255}},  // sky
		{Pos: 0.80, Color: colorista.RGB{R: 80, G: 255, B: 220}},  // mint
		{Pos: 1.00, Color: colorista.RGB{R: 255, G: 240, B: 80}},  // yellow
	}
	sln(sb, "ART\n\n")
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
