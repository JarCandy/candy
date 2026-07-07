package cli

import (
	"os"
	"strings"

	"github.com/JarCandy/candy/pkg/color"
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

	candyGradient := []color.GradientStop{
		{Pos: 0.00, Color: color.RGB{R: 255, G: 80, B: 180}},  // pink
		{Pos: 0.20, Color: color.RGB{R: 255, G: 120, B: 220}}, // candy
		{Pos: 0.40, Color: color.RGB{R: 170, G: 80, B: 255}},  // purple
		{Pos: 0.60, Color: color.RGB{R: 80, G: 180, B: 255}},  // sky
		{Pos: 0.80, Color: color.RGB{R: 80, G: 255, B: 220}},  // mint
		{Pos: 1.00, Color: color.RGB{R: 255, G: 240, B: 80}},  // yellow
	}
	sln(sb, color.Gradient("Usage: candy <command> [options]\n", candyGradient))

	sln(sb, color.Gradient("Commands:", candyGradient))
	s(sb, "  build")
	sln(sb, color.Apply("   Build the project", color.RGBС(color.RGB{R: 217, G: 217, B: 217})))
	s(sb, "  help")
	sln(sb, color.Apply("    Show this help message\n", color.RGBС(color.RGB{R: 217, G: 217, B: 217})))
	sln(sb, color.Gradient("Options:", candyGradient))
	s(sb, "  -h, --help")
	sln(sb, color.Apply("   Show this help message", color.RGBС(color.RGB{R: 217, G: 217, B: 217})))

	println(sb.String())
	return nil
}
