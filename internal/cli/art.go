package cli

import (
	"encoding/base64"
	"encoding/binary"
	"strings"

	"github.com/CandyCrafts/candy/pkg/branding"
	"github.com/rp1s/colorista"
)

const candyArt = `
 ██████╗ █████╗ ███╗   ██╗██████╗ ██╗   ██╗
██╔════╝██╔══██╗████╗  ██║██╔══██╗╚██╗ ██╔╝
██║     ███████║██╔██╗ ██║██║  ██║ ╚████╔╝
██║     ██╔══██║██║╚██╗██║██║  ██║  ╚██╔╝
╚██████╗██║  ██║██║ ╚████║██████╔╝   ██║
 ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝╚═════╝    ╚═╝
`

func Art(color bool) string {
	c := colorista.NewColorista(colorista.ThemeAuto)
	if color {
		return c.Gradient(candyArt, CandyGradientArt) + c.Apply("  Version: [", colorista.Bold) + branding.ReleaseVersion + c.Apply("]", colorista.Bold) + "\n\n"
	}
	return candyArt + c.Apply("  Version: [", colorista.Bold) + branding.ReleaseVersion + c.Apply("]", colorista.Bold) + "\n\n"
}

var CandyGradientArt = []colorista.GradientPos{
	{Pos: 0.00, Color: colorista.RGB{R: 255, G: 70, B: 165}},  // bubblegum
	{Pos: 0.10, Color: colorista.RGB{R: 255, G: 120, B: 220}}, // cotton candy
	{Pos: 0.20, Color: colorista.RGB{R: 210, G: 90, B: 255}},  // grape
	{Pos: 0.30, Color: colorista.RGB{R: 120, G: 90, B: 255}},  // violet blue
	{Pos: 0.40, Color: colorista.RGB{R: 70, G: 170, B: 255}},  // blue raspberry
	{Pos: 0.50, Color: colorista.RGB{R: 60, G: 240, B: 255}},  // ice cyan
	{Pos: 0.60, Color: colorista.RGB{R: 90, G: 255, B: 210}},  // mint
	{Pos: 0.70, Color: colorista.RGB{R: 170, G: 255, B: 130}}, // lime
	{Pos: 0.80, Color: colorista.RGB{R: 255, G: 245, B: 90}},  // lemon
	{Pos: 0.90, Color: colorista.RGB{R: 255, G: 175, B: 90}},  // peach
	{Pos: 1.00, Color: colorista.RGB{R: 255, G: 90, B: 150}},  // strawberry
}

var candyGradient = []colorista.GradientPos{
	{Pos: 0.00, Color: colorista.RGB{R: 255, G: 80, B: 180}},  // pink
	{Pos: 0.20, Color: colorista.RGB{R: 255, G: 120, B: 220}}, // candy
	{Pos: 0.40, Color: colorista.RGB{R: 170, G: 80, B: 255}},  // purple
	{Pos: 0.60, Color: colorista.RGB{R: 80, G: 180, B: 255}},  // sky
	{Pos: 0.80, Color: colorista.RGB{R: 80, G: 255, B: 220}},  // mint
	{Pos: 1.00, Color: colorista.RGB{R: 255, G: 240, B: 80}},  // yellow
}

const Candy = "AD3/ISH/ISED/2Ji/2JiA////////wPb29vb29sD/1JS/1JSA/8nJ/8nJwOwsLCwsLADsAAAsAAAA9kAANkAAAP6Bgb6BgYD/yUl/yUlA/j4+Pj4+AP4AgL4AgID1gAA1gAAA9UAANUAAAP1AAD1AAAD9/f39/f3A9ra2tra2gPxAgLxAgID/wsL/wsLA/cAAPcAAAP2AQH2AQED+gIC+gICA8jIyMjIyAP/FBT/FBQD2dnZ2dnZA64AAK4AAAP/LCz/LCwD/woK/woKA/YAAPYAAAP/WVn/WVkD/AEB/AEBA8gAAMgAAAP/ERH/ERED/05O/05OA/0MDP0MDAP/KSn/KSkD8PDw8PDwA8sAAMsAAAP8/Pz8/PwD1tbW1tbWA9cAANcAAAPzAADzAAADs7Ozs7OzA6phOqphOgONjY2NjY0DmwAAmwAAA+wAAOwAAAP/Fhb/FhYDzMzMzMzMA5tSK5tSKwOsYzysYzwDsWhBsWhBA41EHY1EHQOZUCmZUCkDsGdAsGdAA6lgOalgOQOYTyiYTygDpl02pl02A4pBGopBGgOANxCANxAD//8AAiAgAAAAA+KWiAABAAPilogAAgAJ4paI4paI4paIAAMAA+KWiP//AAIKIAAEAAPilogABQAD4paIAAIAA+KWiAAGAAPilogABwAD4paIAAgAA+KWiAAJAAPilogACgAD4paI//8AAQoAAgAD4paIAAsAA+KWiAAMAAPilogADQAD4paIAA4AA+KWiAAPAAPilogACgAD4paIAAIAA+KWiAAQAAPilogACwAD4paI//8AAQoAAgAG4paI4paIABEAA+KWiAANAAPilogAEgAD4paIABMAA+KWiAAUAAPilogAFQAD4paIABYAA+KWiAAXAAPiloj//wABCgAYAAPilogAGQAD4paIABoAA+KWiAAbAAPilogAHAAD4paIAB0AA+KWiAAeAAPilogAAgAD4paIAB8AA+KWiAAgAAPiloj//wABCgAhAAPilogAHQAD4paIACIAA+KWiAACAAPilogAIwAD4paIACQAA+KWiAACAAPilogAJQAD4paIAAIAA+KWiAAmAAPiloj//wACCiAAAgAG4paI4paIACcAA+KWiAAoAAPilogAKQAD4paIACoAA+KWiAACAAPilogAKwAD4paIACwAA+KWiP//AAMKICAAJwAD4paIAC0AA+KWiAAuAAPilogALwAD4paIADAAA+KWiAAxAAPilogALAAD4paIADIAA+KWiAAzAAPiloj//wAKCiAgICAgICAgIAA0AAPilogANQAD4paIADYAA+KWiAA3AAPiloj//wANCiAgICAgICAgICAgIAAsAAPilogAOAAD4paIADkAA+KWiP//AA4KICAgICAgICAgICAgIAA6AAPilogAOwAD4paIADwAA+KWiA=="
const Amnym = "ALcfHx8AAAADLSEfEAMAAy4hHxIDAAM2LycbEwoDgW5LbVYuA5FeM7t5QQPukGfobTcD511B4TUSA81DLMUnDQOKMyZ6FwgDTiciNgkEA1RCMD0oFAPGp1zpxWwD4NKj/u+5At/czP366ALg3ND++uwC27mW+dKqA+92YOpNMAPrTDDoLg4D4Uw02ygMA3ZrYWFVSQMnJx8KCgADQDsuJiARA9vBafnbdwPf06b98L0C4ODg/v/+AuDg4P7+/gLf2cz99+gC3tC3/OzQAt7Quvzs0wLYtrL1z8oC6WxW5EkuA+DZuv730wKJh31ubF8DMSkjFQwFA8WDQ/WjUwPsYkHnOxID27ut+dTFA9nGxPfh3wKvZUfnhV4D05FH9qlTA82WSfq3WQPZn0z6t1gDzjkzxx4XA9W5tfLSzgLg3bz++9YC4N28//vWAlxcUkZGOgNhLCRMDwYD7Vg+6S0NA+xRNuguDgPYubP10ssC1EA3zSMYA9JIPsghFQPZVkPRNR4D255M+bNWA9KWSPqzVgPPQDjFGhEDw4FC755RA9/Zsv33ygLg3Lv++tUCkXdhspJ3A8hJM74mDQPsSy7pLQwD7oFs6Vo+A8aES/SjXAPcn0z6tVYD0JdK+7dZA9iaSvmyVQPOlkj4tFYD0IlC9KFNA9lmSdJIJgPgwmL+3G8D3sJl/N1zA+Dbu/751AK2aEvviGIDWSgiQgsEAzUiHxoEAAPpUDblLQ0D61U75iwLA9zKqPrlvwLcolD6uFsD2Z9O+7haA9yhTfq3WAPUmkn6tlYD3aJO+7hZA9egTvm5WgPfwWL9228C4MJk/t1yA+DCZP7dcgLf16799MYC73NZ6kgnA3UuI2ISBQM2Ih8bBAAD6FA24ywNA79YTLU9LwPf17H99MkCzpdJ+rhZA9idS/q1VwPyj13uby4D15ZH+a1SA8aKQfesUQPfv2L92W8D38Jk/d1yA+DCY/7dcQLev2H+228DuVsy73ZBA+tMMOgtDQN3LiZlEggDuTorsB8OA8Kag9yvlQPg27j++dEC3KxW+sNiA/SUY/FyMgPCXynxdjMD2qlT+sJfA86BPPSZRwO4YizygDoD3apT+8FeA8qCPfafSwPvaEjqPxYD7E0y6C0MA+1XPegsDANfLCJJDwQDUyUkPAcGA9jOsPbqyAPduGb70XQD7Vw/6TQQA792NvWXRgPfvF/91mwDs0gd7V8nA+tJLugsDAPsUzfoLQwD604x6C4MA+1VOuktDAO7PSmyIwwDJx8fCgAAAy4jIxEFBQPY0bT27cwCwWBB7XZQA+xQNegtDAPIbTHzhTwD38Jj/dxxA9/CY/3dcQLeu1/81WwD0YM99ZpIA+thQ+c+GgO/VTXvakID5Es03ykNA14pI0gMBQNPT0k3NzAD3dm6+/fTAuxjSOhEJAPtUjbpLQwD7Vk46jkSA961XPzOaAPewmT93XID38Jk/dxyA93Mj/voogPf3cz9++gC2Map9uHAA493aXdaSQMuHx8RAAADcG9lWFdLA95WPtcvEgPwel3sVTAD27mU+dKoAt/WtP3zzQLg17b+9M8C3Mal+uG7Ap2MfYFsWAMwJycUCgoDQyQiKQYEA400Jn0YCAPNQCnGJgwD6lI55i0OA+tQNecwEAPfY0rXPB0DoEc1ky4aA2YyKlEWDQM1IyAaBQIDOyMgIAUCAyUfHwcAAAP//wAFICAgICAAAAACOi4AAQABOgACAAEuAAAAAjou//8ABQogICAgAAMAAS0ABAABKgAFAAErAAYAASMABwABKwAIAAErAAkAAT0ACgABPQAAAAE6//8AAwogIAAAAAEuAAsAAS0ADAABKwANAAEgAA4AASAADwABIAAQAAEqABEAASsAEgABPQATAAE9ABQAAS0AFQABOv//AAMKICAAFgABLQAXAAElABgAASAAGQABIAAaAAEgABsAASAAHAABIAAdAAEgAB4AASAAHwABKgAgAAEgACEAASoAAAABOv//AAIKIAAiAAE6ACMAASsAJAABKwAlAAEqACYAASAAJwABPQAoAAEqACkAASsAKgABKgArAAEtACwAASAALQABIAAuAAEgAC8AAS3//wABCgAAAAE6ADAAAT0AMQABKwAyAAEqADMAASAANAABKwA1AAErADYAASoANwABIwA4AAEjADkAASsAOgABKgA7AAEgADwAASAAPQABKwAAAAEt//8AAQoAAAABLgA+AAE9AD8AAT0AQAABKgBBAAErAEIAASoAQwABKwBEAAEqAEUAASsARgABKwBHAAE9AEgAASMASQABKgBKAAEgAEsAAT0ATAABOv//AAEKAE0AAS0ATgABKgBPAAErAFAAASAAUQABIwBSAAEjAFMAASMAVAABIwBVAAEjAFYAASMAVwABIABYAAElAFkAASAAWgABIABbAAEqAFwAAT3//wABCgBdAAEuAF4AASsAXwABPQBgAAEgAGEAASsAYgABKgBjAAErAGQAASoAZQABKwBmAAEjAGcAASoAaAABIABpAAEqAGoAAT0AawABPQBsAAEt//8AAQoAAAABOgBtAAErAG4AASoAbwABIABwAAEjAHEAASMAcgABKwBzAAEjAHQAASoAdQABKgB2AAEjAHcAASMAeAABKgB5AAEqAHoAASsAewABPf//AAEKAAAAAS4AfAABOgB9AAEqAC4AASAAfgABKgB/AAErAIAAAT0AZwABIwCBAAEqAIIAAT0AgwABPQCEAAErAIUAAT0AhgABKwCHAAEtAIgAATr//wACCiAAiQABLQCKAAEgAC0AASAAiwABKwCMAAEqAI0AASsAjgABJQCPAAEgAJAAASUAkQABKgCSAAEqAJMAASsAlAABKgCVAAEt//8AAwogIACWAAE6AJcAASAAmAABPQCZAAErAJoAAT0AmwABKgCcAAEqAJ0AASMAngABKgCfAAEgAKAAASoAoQABKwCiAAEu//8AAwogIAAAAAE6AKMAASsApAABKwB5AAEqADEAASsApQABIwCmAAEgAKcAASAAqAABIACpAAEgAKoAASoAqwABLf//AAUKICAgIACsAAEuAK0AAS0ArgABLQCvAAErALAAAT0AsQABKwCyAAEtALMAAS0AAAABLv//AAYKICAgICAAAAACLToAtAABLQC1AAEtALYAAS0AAAABOg=="
const Confeta = "AGe2JSW2DAwDlwAAnwAAA62trba2tgOvr6+4uLgDn5+fqampA6urq7W1tQPX19fa2toDyEJCxR0dA8cwMMgdHQPeQEDcFhYD7EJC6xwcA9nZ2dnZ2QPJycnKysoD0NDQ0dHRA7u7u8DAwAPIyMjPz88DnwYGqgYGA8YAAMoAAAO8vLy3t7cD9mNj8zAwA/9nZ/82NgPZNjbXEBAD9ktL9SsrA+MyMuQeHgPfJSXfEhID29vb3t7eA68AALYAAAO/AADFAAADzwYG0wEBA9QKCtUAAAPfFhbiDg4Dubm5tra2A+3t7erq6gPk5OTi4uID6+vr6OjoA+U/P+UiIgPuPj7tIiID4Csr4BQUA+ovL+sdHQO8AAC/AAADzAkJzwAAA94XF+EQEAP9MzP9Hx8D7+/v7+/vA9TU1NXV1QPR0dHV1dUDqqqqq6urA93d3dra2gPg4ODe3t4D2DU12BwcA98xMd8UFAPwPDzxJCQD0hkZ0wUFA84PD88AAAPUERHVBgYDyAMDywAAA+IZGeMLCwP+/v7+/v4D4eHh5eXlA9PT09fX1wP///////8D0BQU0wYGA9IxMdAPDwPoQUHnJSUD4TMz4RYWA+IuLuIVFQPoLy/oGxsD4SQk4hQUA/EvL/IbGwPmIiLoGBgDyAEBywAAA+EXF+IKCgPMzMzQ0NADtLS0urq6A8jIyMzMzAPu7u7u7u4D1RMT1gcHA9EhIdEGBgPoNDToHBwD8jc38R0dA+IjI+INDQPrKCjqDg4D1NTU1NTUA8cBAcsAAAPiGRnmFBQD2xIS3w0NA7gAAL4AAAPHAADLAAAD8vLy8vLyA+AcHOMSEgP3Pj73GBgD6enp6OjoA+vr6+zs7AP19fX29vYDxcXFyMjIA80KCtMKCgP3Ojr3FhYD/1BQ/ywsA8jIyMvLywP09PT09PQD+vr6+vr6A/I1NfIREQP8Pz/8GxsD//8ABCAgICAAAAAD4paRAAEAA+KWkQACAAPilpMAAwAD4paTAAQAA+KWkwAFAAPilpMABgAD4paT//8AAwogIAAHAAPilpEACAAD4paRAAkAA+KWkQAKAAPilpEACwAD4paTAAwAA+KWkwANAAPilpMADgAD4paTAA8AA+KWkwAQAAPilpEAEQAD4paR//8AAgogABIAA+KWkwATAAPilpIAFAAD4paSABUAA+KWkQAWAAPilpIAFwAD4paRABgAA+KWkQAZAAPilpMAGgAD4paRABsAA+KWkQAcAAPilpEAHQAD4paRAB4AA+KWkf//AAEKAB8AA+KWkwAgAAPilogAIQAD4paIACIAA+KWiAAjAAPilpEAJAAD4paRACUAA+KWkQAmAAPilpEAJwAD4paRACgAA+KWkQApAAPilpEAKgAD4paSACsAA+KWiAAsAAPilpMALQAD4paT//8AAQoALgAD4paTAC8AA+KWkwAwAAPilpMAMQAD4paRADIAA+KWkQAzAAPilpIANAAD4paRADUAA+KWkQA2AAPilpEANwAD4paRADgAA+KWkQA5AAPilogAOgAD4paIADsAA+KWkwA8AAPiloj//wABCgA9AAPilpEAPgAD4paRAD8AA+KWkQBAAAPilpEAQQAD4paRAEIAA+KWkQBDAAPilpEARAAD4paRAEUAA+KWkQBGAAPilpEARwAD4paRAEgAA+KWkwBJAAPilpMASgAD4paTAEsAA+KWiP//AAIKIABMAAPilpEATQAD4paRAE4AA+KWkQBPAAPilpEAUAAD4paRAFEAA+KWkQBSAAPilpMAUwAD4paRAFQAA+KWkQBVAAPilpEAVgAD4paRAFcAA+KWkQBYAAPiloj//wADCiAgAFkAA+KWkQBaAAPilpEAWwAD4paIAFwAA+KWiABdAAPilogAXgAD4paTAC0AA+KWkwBfAAPilpEAVwAD4paRAGAAA+KWkQBhAAPilpL//wAFCiAgICAAYgAD4paTAFgAA+KWiABjAAPilogAZAAD4paIADwAA+KWiABlAAPilpEAZgAD4paR"

// RenderString decodes a compact Rigel base64 payload and returns one colored terminal string.
func RenderString(payload string) string {
	c := colorista.NewColorista(colorista.ThemeAuto)
	_ = c
	raw, _ := base64.StdEncoding.DecodeString(payload)
	offset := 0
	readUint16 := func() uint16 {
		value := binary.BigEndian.Uint16(raw[offset:])
		offset += 2
		return value
	}
	styleCount := int(readUint16())
	type renderStyle struct {
		fg, bg colorista.RGB
		flags  byte
	}
	styles := make([]renderStyle, styleCount)
	for i := range styles {
		styles[i].fg = colorista.RGB{R: raw[offset], G: raw[offset+1], B: raw[offset+2]}
		styles[i].bg = colorista.RGB{R: raw[offset+3], G: raw[offset+4], B: raw[offset+5]}
		styles[i].flags = raw[offset+6]
		offset += 7
	}
	var out strings.Builder
	for offset < len(raw) {
		styleIndex := readUint16()
		length := int(readUint16())
		text := string(raw[offset : offset+length])
		offset += length
		if styleIndex == 0xffff {
			out.WriteString(text)
			continue
		}
		style := styles[int(styleIndex)]
		switch style.flags {
		case 1:
			out.WriteString(c.Apply(text, colorista.Rgb(style.fg)))
		case 2:
			out.WriteString(c.Apply(text, colorista.BgRgb(style.bg)))
		default:
			out.WriteString(c.Apply(text, colorista.Rgb(style.fg), colorista.BgRgb(style.bg)))
		}
	}
	return out.String()
}
