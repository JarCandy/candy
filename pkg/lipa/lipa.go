// Package lipa renders Go values as an interactive minimal tree.
package lipa

import (
	"os"
	"path/filepath"
)

type Option func(*Options)

type Options struct {
	Title        string
	MaxDepth     int
	MaxNodes     int
	ExpandDepth  int
	OutputPath   string
	OpenBrowser  bool
	HiddenFields []string
	ShowHidden   bool
	Source       string
}

func defaultOptions() Options {
	return Options{
		Title:       "Lipa Tree",
		MaxDepth:    18,
		MaxNodes:    4000,
		ExpandDepth: 4,
		OpenBrowser: true,
	}
}

func WithTitle(title string) Option {
	return func(o *Options) {
		o.Title = title
	}
}

func WithMaxDepth(depth int) Option {
	return func(o *Options) {
		o.MaxDepth = depth
	}
}

func WithMaxNodes(nodes int) Option {
	return func(o *Options) {
		o.MaxNodes = nodes
	}
}

func WithExpandDepth(depth int) Option {
	return func(o *Options) {
		o.ExpandDepth = depth
	}
}

func WithOutputPath(path string) Option {
	return func(o *Options) {
		o.OutputPath = path
	}
}

func WithHiddenFields(names ...string) Option {
	return func(o *Options) {
		o.HiddenFields = append(o.HiddenFields, names...)
	}
}

func WithShowHidden() Option {
	return func(o *Options) {
		o.ShowHidden = true
	}
}

func WithSource(source string) Option {
	return func(o *Options) {
		o.Source = source
	}
}

func WithoutOpen() Option {
	return func(o *Options) {
		o.OpenBrowser = false
	}
}

func View(value any, opts ...Option) error {
	html, options, err := render(value, opts...)
	if err != nil {
		return err
	}

	path := options.OutputPath
	if path == "" {
		file, err := os.CreateTemp("", "lipa-*.html")
		if err != nil {
			return err
		}
		path = file.Name()
		if _, err := file.WriteString(html); err != nil {
			_ = file.Close()
			return err
		}
		if err := file.Close(); err != nil {
			return err
		}
	} else {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(html), 0o644); err != nil {
			return err
		}
	}

	if options.OpenBrowser {
		return openBrowser(path)
	}
	return nil
}

func Render(value any, opts ...Option) (string, error) {
	html, _, err := render(value, opts...)
	return html, err
}

func Build(value any, opts ...Option) *Node {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return buildTree(value, options)
}

func render(value any, opts ...Option) (string, Options, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	root := buildTree(value, options)
	return renderHTML(root, options), options, nil
}
