package color

import (
	"fmt"
	"strings"
)

var Enabled = true

const Reset = "\033[0m"

type Style string

// Text colors
const (
	Black   Style = "30"
	Red     Style = "31"
	Green   Style = "32"
	Yellow  Style = "33"
	Blue    Style = "34"
	Magenta Style = "35"
	Cyan    Style = "36"
	White   Style = "37"

	BrightBlack   Style = "90"
	BrightRed     Style = "91"
	BrightGreen   Style = "92"
	BrightYellow  Style = "93"
	BrightBlue    Style = "94"
	BrightMagenta Style = "95"
	BrightCyan    Style = "96"
	BrightWhite   Style = "97"
)

// Background colors
const (
	BgBlack   Style = "40"
	BgRed     Style = "41"
	BgGreen   Style = "42"
	BgYellow  Style = "43"
	BgBlue    Style = "44"
	BgMagenta Style = "45"
	BgCyan    Style = "46"
	BgWhite   Style = "47"
)

// Effects
const (
	ResetStyle Style = "0"

	Bold          Style = "1"
	Faint         Style = "2"
	Italic        Style = "3"
	Underline     Style = "4"
	SlowBlink     Style = "5"
	RapidBlink    Style = "6"
	Reverse       Style = "7"
	Hidden        Style = "8"
	StrikeThrough Style = "9"

	DoubleUnderline Style = "21"

	NormalIntensity Style = "22"
	NoItalic        Style = "23"
	NoUnderline     Style = "24"
	NoBlink         Style = "25"
	NoReverse       Style = "27"
	Reveal          Style = "28"
	NoStrike        Style = "29"

	Frame    Style = "51"
	Encircle Style = "52"
	Overline Style = "53"

	NoFrame    Style = "54"
	NoOverline Style = "55"
)

func Apply(text any, styles ...Style) string {
	s := fmt.Sprint(text)

	if !Enabled || len(styles) == 0 {
		return s
	}

	codes := make([]string, len(styles))
	for i, st := range styles {
		codes[i] = string(st)
	}

	return "\033[" + strings.Join(codes, ";") + "m" + s + Reset
}

type RGB struct {
	R, G, B uint8
}

func RGBС(color RGB) Style {
	return Style(fmt.Sprintf("38;2;%d;%d;%d", color.R, color.G, color.B))
}

func BgRGB(color RGB) Style {
	return Style(fmt.Sprintf("48;2;%d;%d;%d", color.R, color.G, color.B))
}

func ANSI256(i uint8) Style {
	return Style(fmt.Sprintf("38;5;%d", i))
}

func BgANSI256(i uint8) Style {
	return Style(fmt.Sprintf("48;5;%d", i))
}

//

func Success(a any) string {
	return Apply(a, Bold, BrightGreen)
}

func Error(a any) string {
	return Apply(a, Bold, BrightRed)
}

func Warning(a any) string {
	return Apply(a, Bold, BrightYellow)
}

func Info(a any) string {
	return Apply(a, Bold, BrightBlue)
}

func Debug(a any) string {
	return Apply(a, BrightBlack)
}

func Title(a any) string {
	return Apply(a, Bold, Underline, BrightWhite)
}

type GradientStop struct {
	Pos   float64 // 0.0 - 1.0
	Color RGB
}

func rgbStyle(c RGB) Style {
	return Style(fmt.Sprintf("38;2;%d;%d;%d", c.R, c.G, c.B))
}

func bgRGBStyle(c RGB) Style {
	return Style(fmt.Sprintf("48;2;%d;%d;%d", c.R, c.G, c.B))
}

func lerp(a, b uint8, t float64) uint8 {
	return uint8(float64(a) + (float64(b)-float64(a))*t)
}

func interpolate(a, b RGB, t float64) RGB {
	return RGB{
		R: lerp(a.R, b.R, t),
		G: lerp(a.G, b.G, t),
		B: lerp(a.B, b.B, t),
	}
}

func colorAt(stops []GradientStop, pos float64) RGB {
	if len(stops) == 0 {
		return RGB{}
	}

	if pos <= stops[0].Pos {
		return stops[0].Color
	}

	for i := 0; i < len(stops)-1; i++ {
		a := stops[i]
		b := stops[i+1]

		if pos >= a.Pos && pos <= b.Pos {
			t := (pos - a.Pos) / (b.Pos - a.Pos)
			return interpolate(a.Color, b.Color, t)
		}
	}

	return stops[len(stops)-1].Color
}

func Gradient(text string, stops []GradientStop, styles ...Style) string {
	if !Enabled {
		return text
	}

	r := []rune(text)

	var out strings.Builder

	for i, ch := range r {

		p := 0.0
		if len(r) > 1 {
			p = float64(i) / float64(len(r)-1)
		}

		c := colorAt(stops, p)

		s := append([]Style{}, styles...)
		s = append(s, rgbStyle(c))

		out.WriteString(Apply(string(ch), s...))
	}

	return out.String()
}

func BgGradient(text string, stops []GradientStop, styles ...Style) string {
	if !Enabled {
		return text
	}

	r := []rune(text)

	var out strings.Builder

	for i, ch := range r {

		p := 0.0
		if len(r) > 1 {
			p = float64(i) / float64(len(r)-1)
		}

		c := colorAt(stops, p)

		s := append([]Style{}, styles...)
		s = append(s, bgRGBStyle(c))

		out.WriteString(Apply(string(ch), s...))
	}

	return out.String()
}

func FullGradient(
	text string,
	fg []GradientStop,
	bg []GradientStop,
	styles ...Style,
) string {
	if !Enabled {
		return text
	}

	runes := []rune(text)

	var out strings.Builder

	for i, ch := range runes {
		pos := 0.0

		if len(runes) > 1 {
			pos = float64(i) / float64(len(runes)-1)
		}

		codes := make([]Style, 0, len(styles)+2)

		codes = append(codes, styles...)

		if len(fg) > 0 {
			fgColor := colorAt(fg, pos)
			codes = append(codes, rgbStyle(fgColor))
		}

		if len(bg) > 0 {
			bgColor := colorAt(bg, pos)
			codes = append(codes, bgRGBStyle(bgColor))
		}

		out.WriteString(Apply(string(ch), codes...))
	}

	return out.String()
}
