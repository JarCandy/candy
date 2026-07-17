package clifmt

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/CandyCrafts/candy/internal/database"
	"github.com/rp1s/colorista"
	"github.com/rp1s/digreyt/translate"
)

type Text = translate.Translations

type Document struct {
	Art      string
	ShowArt  bool
	Title    Text
	Usage    Text
	Sections []Section
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
	Language   string
	Color      *colorista.Colorista
	Auto       bool
	Timeout    time.Duration
	cache      map[string]string
	cacheStore database.CLITextCache

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

var defaultCacheStore database.CLITextCache

func SetDefaultCacheStore(store database.CLITextCache) {
	defaultCacheStore = store
}

func New(language string) *Renderer {
	return &Renderer{
		Language:     language,
		Color:        colorista.NewColorista(colorista.ThemeAuto),
		Auto:         true,
		Timeout:      2 * time.Second,
		TitleColor:   colorista.RGB{R: 255, G: 96, B: 190},
		SectionColor: colorista.RGB{R: 96, G: 210, B: 255},
		LabelColor:   colorista.RGB{R: 245, G: 245, B: 245},
		TextColor:    colorista.RGB{R: 218, G: 218, B: 218},
		MutedColor:   colorista.RGB{R: 150, G: 150, B: 150},
	}
}

func (self *Renderer) Render(doc Document) string {
	self.ensureDefaults()

	var sb strings.Builder
	if doc.ShowArt && doc.Art != "" {
		sb.WriteString(doc.Art)
		if !strings.HasSuffix(doc.Art, "\n") {
			sb.WriteString("\n")
		}
	}

	if title := self.text(doc.Title); title != "" {
		sb.WriteString(self.title(title))
		sb.WriteString("\n")
	}
	if usage := self.text(doc.Usage); usage != "" {
		sb.WriteString(self.muted(self.text(T("Usage:"))))
		sb.WriteString("\n  ")
		sb.WriteString(self.normal(usage))
		sb.WriteString("\n")
	}

	for _, section := range doc.Sections {
		self.renderSection(&sb, section)
	}

	return strings.TrimRight(sb.String(), "\n") + "\n"
}

func (self *Renderer) renderSection(sb *strings.Builder, section Section) {
	title := self.text(section.Title)
	if title == "" && len(section.Rows) == 0 {
		return
	}

	sb.WriteString("\n")
	if title != "" {
		sb.WriteString(self.section(title + ":"))
		sb.WriteString("\n")
	}

	for _, row := range section.Rows {
		self.renderRow(sb, row, 2)
	}
}

func (self *Renderer) renderRow(sb *strings.Builder, row Row, indent int) {
	label := strings.TrimSpace(row.Label)
	desc := self.text(row.Description)
	if label == "" && desc == "" {
		return
	}

	prefix := strings.Repeat(" ", indent)
	if label == "" {
		sb.WriteString(prefix)
		sb.WriteString(self.normal(desc))
		sb.WriteString("\n")
		return
	}

	sb.WriteString(prefix)
	sb.WriteString(self.label(label))
	if desc != "" {
		sb.WriteString("\n")
		sb.WriteString(prefix)
		sb.WriteString("  ")
		sb.WriteString(self.normal(desc))
	}
	sb.WriteString("\n")

	if len(row.Children) > 0 {
		childWidth := maxLabelWidth(row.Children, 0)
		for _, child := range row.Children {
			self.renderChildRow(sb, child, childWidth, indent+4)
		}
	}
}

func (self *Renderer) renderChildRow(sb *strings.Builder, row Row, width int, indent int) {
	label := strings.TrimSpace(row.Label)
	desc := self.text(row.Description)
	if label == "" && desc == "" {
		return
	}

	prefix := strings.Repeat(" ", indent)
	if label == "" {
		sb.WriteString(prefix)
		sb.WriteString(self.normal(desc))
		sb.WriteString("\n")
		return
	}

	sb.WriteString(prefix)
	sb.WriteString(self.label(label))
	if desc != "" {
		padding := width - len([]rune(label))
		if padding < 1 {
			padding = 1
		}
		sb.WriteString(strings.Repeat(" ", padding+2))
		sb.WriteString(self.normal(desc))
	}
	sb.WriteString("\n")
}

func (self *Renderer) text(values Text) string {
	if !self.Auto || hasLanguage(self.Language, values) || normalizeLanguage(self.Language) == normalizeLanguage(translate.DefaultLanguage) {
		return translate.ResolveFor(self.Language, translate.Translations(values))
	}

	key := self.cacheKey(values)
	if cached, ok := self.cache[key]; ok {
		return cached
	}

	if self.cacheStore != nil {
		ctx, cancel := context.WithTimeout(context.Background(), self.Timeout)
		defer cancel()
		if cachedText, ok, err := self.cacheStore.ReadText(ctx, self.originalText(values), self.Language); err == nil && ok {
			self.cache[key] = cachedText
			return cachedText
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), self.Timeout)
	defer cancel()

	text, err := translate.ResolveAutoFor(ctx, self.Language, translate.Translations(values))
	if err != nil || isAutoTranslationFailure(text) {
		text = translate.ResolveFor(self.Language, translate.Translations(values))
	}
	self.cache[key] = text
	if self.cacheStore != nil {
		ctx2, cancel2 := context.WithTimeout(context.Background(), self.Timeout)
		defer cancel2()
		if err := self.cacheStore.WriteText(ctx2, self.originalText(values), self.Language, text); err != nil {
			log.Printf("clifmt: failed to write translation cache: %v", err)
		}
	}
	return text
}

func (self *Renderer) label(text string) string {
	return self.Color.Apply(text, colorista.Bold, colorista.Rgb(self.LabelColor))
}

func (self *Renderer) normal(text string) string {
	return self.Color.Apply(text, colorista.Rgb(self.TextColor))
}

func (self *Renderer) muted(text string) string {
	return self.Color.Apply(text, colorista.Rgb(self.MutedColor))
}

func (self *Renderer) title(text string) string {
	return self.Color.Apply(text, colorista.Bold, colorista.Rgb(self.TitleColor))
}

func (self *Renderer) section(text string) string {
	return self.Color.Apply(text, colorista.Bold, colorista.Rgb(self.SectionColor))
}

func (self *Renderer) ensureDefaults() {
	if self.Color == nil {
		self.Color = colorista.NewColorista(colorista.ThemeAuto)
	}
	if strings.TrimSpace(self.Language) == "" {
		self.Language = translate.DefaultLanguage
	}
	if self.Timeout <= 0 {
		self.Timeout = 2 * time.Second
	}
	if self.cache == nil {
		self.cache = make(map[string]string)
	}
	if self.cacheStore == nil {
		self.cacheStore = defaultCacheStore
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

func (self *Renderer) cacheKey(values Text) string {
	var sb strings.Builder
	sb.WriteString(normalizeLanguage(self.Language))
	for _, value := range values {
		sb.WriteString("\x00")
		sb.WriteString(normalizeLanguage(value.Language))
		sb.WriteString("=")
		sb.WriteString(value.Text)
	}
	return sb.String()
}

func (self *Renderer) cacheEntryKey(values Text) string {
	return fmt.Sprintf("%s:%s", normalizeLanguage(self.Language), self.originalText(values))
}

func (self *Renderer) originalText(values Text) string {
	for _, value := range values {
		if normalizeLanguage(value.Language) == normalizeLanguage(translate.DefaultLanguage) {
			return value.Text
		}
	}
	if len(values) == 0 {
		return ""
	}
	return values[0].Text
}

func hasLanguage(language string, values Text) bool {
	language = normalizeLanguage(language)
	for _, value := range values {
		if normalizeLanguage(value.Language) == language {
			return true
		}
	}
	return false
}

func normalizeLanguage(language string) string {
	return strings.ToLower(strings.TrimSpace(language))
}

func isAutoTranslationFailure(text string) bool {
	normalized := strings.ToLower(strings.TrimSpace(text))
	if normalized == "" {
		return true
	}

	for _, marker := range []string{
		"invalid target language",
		"invalid source language",
		"langpair=",
		"no query specified",
	} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}
