package clifmt

import (
	"fmt"
	"strings"

	"github.com/rp1s/colorista"
	"github.com/rp1s/digreyt/translate"
)

type Text = translate.Translations

type Document struct {
	Art         string
	ArtGradient []colorista.GradientPos
	Title       Text
	Usage       Text
	Sections    []Section
}

type Section struct {
	Title Text
	Rows  []Row
}

type Row struct {
	Label       string
	Description Text
	Children    []Row
}

type Renderer struct {
	Language string
	Color    *colorista.Colorista

	TitleColor   colorista.RGB
	SectionColor colorista.RGB
	LabelColor   colorista.RGB
	TextColor    colorista.RGB
	MutedColor   colorista.RGB
}

func T(eng string, values ...translate.Translation) Text {
	out := Text{{Language: translate.DefaultLanguage, Text: eng}}
	out = append(out, values...)
	return out
}

func Lang(language string, text string) translate.Translation {
	return translate.Translation{Language: language, Text: text}
}

func New(language string) *Renderer {
	return &Renderer{
		Language:     language,
		Color:        colorista.NewColorista(colorista.ThemeAuto),
		TitleColor:   colorista.RGB{R: 255, G: 96, B: 190},
		SectionColor: colorista.RGB{R: 96, G: 210, B: 255},
		LabelColor:   colorista.RGB{R: 245, G: 245, B: 245},
		TextColor:    colorista.RGB{R: 218, G: 218, B: 218},
		MutedColor:   colorista.RGB{R: 150, G: 150, B: 150},
	}
}

func (r *Renderer) Render(doc Document) string {
	r.ensureDefaults()

	var sb strings.Builder
	if doc.Art != "" {
		if len(doc.ArtGradient) > 0 {
			sb.WriteString(r.Color.Gradient(doc.Art, doc.ArtGradient))
		} else {
			sb.WriteString(doc.Art)
		}
		if !strings.HasSuffix(doc.Art, "\n") {
			sb.WriteString("\n")
		}
	}

	if title := r.text(doc.Title); title != "" {
		sb.WriteString(r.title(title))
		sb.WriteString("\n")
	}
	if usage := r.text(doc.Usage); usage != "" {
		sb.WriteString(r.muted("Usage:"))
		sb.WriteString("\n  ")
		sb.WriteString(r.normal(usage))
		sb.WriteString("\n")
	}

	for _, section := range doc.Sections {
		r.renderSection(&sb, section)
	}

	return strings.TrimRight(sb.String(), "\n") + "\n"
}

func (r *Renderer) renderSection(sb *strings.Builder, section Section) {
	title := r.text(section.Title)
	if title == "" && len(section.Rows) == 0 {
		return
	}

	sb.WriteString("\n")
	if title != "" {
		sb.WriteString(r.section(title + ":"))
		sb.WriteString("\n")
	}

	for _, row := range section.Rows {
		r.renderRow(sb, row, 2)
	}
}

func (r *Renderer) renderRow(sb *strings.Builder, row Row, indent int) {
	label := strings.TrimSpace(row.Label)
	desc := r.text(row.Description)
	if label == "" && desc == "" {
		return
	}

	prefix := strings.Repeat(" ", indent)
	if label == "" {
		sb.WriteString(prefix)
		sb.WriteString(r.normal(desc))
		sb.WriteString("\n")
		return
	}

	sb.WriteString(prefix)
	sb.WriteString(r.label(label))
	if desc != "" {
		sb.WriteString("\n")
		sb.WriteString(prefix)
		sb.WriteString("  ")
		sb.WriteString(r.normal(desc))
	}
	sb.WriteString("\n")

	if len(row.Children) > 0 {
		childWidth := maxLabelWidth(row.Children, 0)
		for _, child := range row.Children {
			r.renderChildRow(sb, child, childWidth, indent+4)
		}
	}
}

func (r *Renderer) renderChildRow(sb *strings.Builder, row Row, width int, indent int) {
	label := strings.TrimSpace(row.Label)
	desc := r.text(row.Description)
	if label == "" && desc == "" {
		return
	}

	prefix := strings.Repeat(" ", indent)
	if label == "" {
		sb.WriteString(prefix)
		sb.WriteString(r.normal(desc))
		sb.WriteString("\n")
		return
	}

	sb.WriteString(prefix)
	sb.WriteString(r.label(label))
	if desc != "" {
		padding := width - len([]rune(label))
		if padding < 1 {
			padding = 1
		}
		sb.WriteString(strings.Repeat(" ", padding+2))
		sb.WriteString(r.normal(desc))
	}
	sb.WriteString("\n")
}

func (r *Renderer) text(values Text) string {
	return translate.ResolveFor(r.Language, translate.Translations(values))
}

func (r *Renderer) label(text string) string {
	return r.Color.Apply(text, colorista.Bold, colorista.Rgb(r.LabelColor))
}

func (r *Renderer) normal(text string) string {
	return r.Color.Apply(text, colorista.Rgb(r.TextColor))
}

func (r *Renderer) muted(text string) string {
	return r.Color.Apply(text, colorista.Rgb(r.MutedColor))
}

func (r *Renderer) title(text string) string {
	return r.Color.Apply(text, colorista.Bold, colorista.Rgb(r.TitleColor))
}

func (r *Renderer) section(text string) string {
	return r.Color.Apply(text, colorista.Bold, colorista.Rgb(r.SectionColor))
}

func (r *Renderer) ensureDefaults() {
	if r.Color == nil {
		r.Color = colorista.NewColorista(colorista.ThemeAuto)
	}
	if strings.TrimSpace(r.Language) == "" {
		r.Language = translate.DefaultLanguage
	}
}

func maxLabelWidth(rows []Row, fallback int) int {
	width := fallback
	for _, row := range rows {
		if size := len([]rune(row.Label)); size > width {
			width = size
		}
	}
	return width
}

func Sprint(doc Document, language string) string {
	return New(language).Render(doc)
}

func Print(doc Document, language string) {
	fmt.Print(Sprint(doc, language))
}
