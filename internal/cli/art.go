package cli

import (
	"github.com/caramelang/caramel/pkg/branding"
	"github.com/rp1s/colorista"
)

const caramelArt = `
 ██████╗ █████╗ ██████╗  █████╗ ███╗   ███╗███████╗██╗
██╔════╝██╔══██╗██╔══██╗██╔══██╗████╗ ████║██╔════╝██║
██║     ███████║██████╔╝███████║██╔████╔██║█████╗  ██║
██║     ██╔══██║██╔══██╗██╔══██║██║╚██╔╝██║██╔══╝  ██║
╚██████╗██║  ██║██║  ██║██║  ██║██║ ╚═╝ ██║███████╗███████╗
 ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝╚══════╝╚══════╝
`

func Art(color bool) string {
	cls := colorista.NewColorista(colorista.ThemeAuto)
	version := cls.Apply("  Version: [", colorista.Bold) +
		branding.ReleaseVersion +
		cls.Apply("]", colorista.Bold) +
		"\n\n"
	if color {
		return cls.Gradient(caramelArt, caramelGradientArt) + version
	}
	return caramelArt + version
}

var caramelGradientArt = []colorista.GradientPos{
	{Pos: 0.00, Color: colorista.RGB{R: 255, G: 70, B: 165}},
	{Pos: 0.10, Color: colorista.RGB{R: 255, G: 120, B: 220}},
	{Pos: 0.20, Color: colorista.RGB{R: 210, G: 90, B: 255}},
	{Pos: 0.30, Color: colorista.RGB{R: 120, G: 90, B: 255}},
	{Pos: 0.40, Color: colorista.RGB{R: 70, G: 170, B: 255}},
	{Pos: 0.50, Color: colorista.RGB{R: 60, G: 240, B: 255}},
	{Pos: 0.60, Color: colorista.RGB{R: 90, G: 255, B: 210}},
	{Pos: 0.70, Color: colorista.RGB{R: 170, G: 255, B: 130}},
	{Pos: 0.80, Color: colorista.RGB{R: 255, G: 245, B: 90}},
	{Pos: 0.90, Color: colorista.RGB{R: 255, G: 175, B: 90}},
	{Pos: 1.00, Color: colorista.RGB{R: 255, G: 90, B: 150}},
}
