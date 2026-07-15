package lipa

type Node struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Kind     string         `json:"kind"`
	Type     string         `json:"type"`
	Value    string         `json:"value,omitempty"`
	Ref      string         `json:"ref,omitempty"`
	Snippet  *SourceSnippet `json:"snippet,omitempty"`
	Nil      bool           `json:"nil,omitempty"`
	Cycle    bool           `json:"cycle,omitempty"`
	Trunc    bool           `json:"trunc,omitempty"`
	Hidden   bool           `json:"hidden,omitempty"`
	Children []*Node        `json:"children,omitempty"`
}

type SourceSnippet struct {
	FileName string `json:"fileName,omitempty"`
	Line     uint64 `json:"line,omitempty"`
	Column   uint64 `json:"column,omitempty"`
	Text     string `json:"text,omitempty"`
	Marker   string `json:"marker,omitempty"`
}
