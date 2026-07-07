package main

import (
	"github.com/JarCandy/candy/core/cli"
	"github.com/JarCandy/candy/pkg/color"
)

func main() {
	color.SetTheme(color.DetectTheme())
	cli.HandlerCmd()
}
