package lipa

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type treeBuilder struct {
	options      Options
	hiddenFields map[string]bool
	seen         map[visitKey]string
	nextID       int
	nodes        int
}

type visitKey struct {
	typ reflect.Type
	ptr uintptr
}

func buildTree(value any, options Options) *Node {
	if options.Source == "" {
		options.Source = sourceField(reflect.ValueOf(value))
	}
	builder := &treeBuilder{
		options:      options,
		hiddenFields: makeHiddenFields(options.HiddenFields),
		seen:         make(map[visitKey]string),
	}
	return builder.value("root", reflect.ValueOf(value), 0)
}

func (self *treeBuilder) value(name string, v reflect.Value, depth int) *Node {
	id := self.id()
	node := &Node{ID: id, Name: name}

	if !v.IsValid() {
		node.Kind = "nil"
		node.Value = "nil"
		node.Nil = true
		return node
	}

	node.Kind = v.Kind().String()
	node.Type = v.Type().String()
	node.Snippet = self.sourceSnippet(v)

	if self.options.MaxNodes > 0 && self.nodes >= self.options.MaxNodes {
		node.Value = "max nodes reached"
		node.Trunc = true
		return node
	}
	self.nodes++

	if self.options.MaxDepth > 0 && depth >= self.options.MaxDepth {
		node.Value = "max depth reached"
		node.Trunc = true
		return node
	}

	switch v.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			node.Nil = true
			node.Value = "nil"
			return node
		}
		return self.value(name, v.Elem(), depth)

	case reflect.Pointer:
		if v.IsNil() {
			node.Nil = true
			node.Value = "nil"
			return node
		}
		key := visitKey{typ: v.Type(), ptr: v.Pointer()}
		if oldID, ok := self.seen[key]; ok {
			node.Cycle = true
			node.Ref = oldID
			node.Value = "cycle -> " + oldID
			return node
		}
		self.seen[key] = node.ID
		inner := self.value(name, v.Elem(), depth)
		inner.ID = node.ID
		return inner

	case reflect.Struct:
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" && !field.Anonymous {
				node.Children = append(node.Children, &Node{
					ID:     self.id(),
					Name:   field.Name,
					Kind:   v.Field(i).Kind().String(),
					Type:   field.Type.String(),
					Value:  "<unexported>",
					Hidden: self.isHiddenField(field.Name),
				})
				continue
			}
			child := self.value(field.Name, v.Field(i), depth+1)
			child.Hidden = self.isHiddenField(field.Name)
			node.Children = append(node.Children, child)
		}

	case reflect.Slice, reflect.Array:
		node.Value = fmt.Sprintf("len=%d", v.Len())
		if v.Kind() == reflect.Slice {
			node.Value += fmt.Sprintf(" cap=%d", v.Cap())
		}
		for i := 0; i < v.Len(); i++ {
			node.Children = append(node.Children, self.value("["+strconv.Itoa(i)+"]", v.Index(i), depth+1))
		}

	case reflect.Map:
		if v.IsNil() {
			node.Nil = true
			node.Value = "nil"
			return node
		}
		node.Value = fmt.Sprintf("len=%d", v.Len())
		keys := v.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return fmt.Sprint(valueInterface(keys[i])) < fmt.Sprint(valueInterface(keys[j]))
		})
		for _, key := range keys {
			child := self.value(fmt.Sprintf("[%v]", valueInterface(key)), v.MapIndex(key), depth+1)
			child.Children = append([]*Node{self.value("key", key, depth+1)}, child.Children...)
			node.Children = append(node.Children, child)
		}

	case reflect.String:
		node.Value = quoteString(v.String())

	case reflect.Bool:
		node.Value = scalarString(v, strconv.FormatBool(v.Bool()))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		node.Value = scalarString(v, strconv.FormatInt(v.Int(), 10))

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		node.Value = scalarString(v, strconv.FormatUint(v.Uint(), 10))

	case reflect.Float32, reflect.Float64:
		node.Value = scalarString(v, strconv.FormatFloat(v.Float(), 'g', -1, v.Type().Bits()))

	case reflect.Complex64, reflect.Complex128:
		node.Value = scalarString(v, fmt.Sprint(v.Complex()))

	case reflect.Func:
		if v.IsNil() {
			node.Nil = true
			node.Value = "nil"
		} else {
			node.Value = "func"
		}

	case reflect.Chan:
		if v.IsNil() {
			node.Nil = true
			node.Value = "nil"
		} else {
			node.Value = fmt.Sprintf("len=%d cap=%d", v.Len(), v.Cap())
		}

	default:
		node.Value = fmt.Sprint(valueInterface(v))
	}

	return node
}

func makeHiddenFields(names []string) map[string]bool {
	fields := make(map[string]bool, len(names))
	for _, name := range names {
		fields[name] = true
	}
	return fields
}

func (self *treeBuilder) isHiddenField(name string) bool {
	return self.hiddenFields[name]
}

func sourceField(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}
	for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return ""
	}
	field := v.FieldByName("Source")
	if !field.IsValid() {
		return ""
	}
	for field.Kind() == reflect.Interface || field.Kind() == reflect.Pointer {
		if field.IsNil() {
			return ""
		}
		field = field.Elem()
	}
	if field.Kind() != reflect.String {
		return ""
	}
	return field.String()
}

func (self *treeBuilder) sourceSnippet(v reflect.Value) *SourceSnippet {
	if self.options.Source == "" {
		return nil
	}
	pos, ok := extractPosition(v)
	if !ok || pos.Line == 0 || pos.Column == 0 {
		return nil
	}
	return buildSourceSnippet(self.options.Source, pos)
}

type sourcePosition struct {
	FileName string
	Line     uint64
	Column   uint64
	Offset   uint64
}

func extractPosition(v reflect.Value) (sourcePosition, bool) {
	return extractPositionDepth(v, 0)
}

func extractPositionDepth(v reflect.Value, depth int) (sourcePosition, bool) {
	if depth > 6 || !v.IsValid() {
		return sourcePosition{}, false
	}
	for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return sourcePosition{}, false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return sourcePosition{}, false
	}

	if line, ok := uintField(v, "Line"); ok && line > 0 {
		column, _ := uintField(v, "Column")
		fileName, _ := stringField(v, "FileName")
		offset, _ := uintField(v, "Offset")
		return sourcePosition{
			FileName: fileName,
			Line:     line,
			Column:   column,
			Offset:   offset,
		}, true
	}

	for _, name := range []string{"Pos", "Tok", "Tok_s", "Tok_e"} {
		field := v.FieldByName(name)
		if !field.IsValid() {
			continue
		}
		if pos, ok := extractPositionDepth(field, depth+1); ok {
			return pos, true
		}
	}

	return sourcePosition{}, false
}

func uintField(v reflect.Value, name string) (uint64, bool) {
	field := v.FieldByName(name)
	if !field.IsValid() {
		return 0, false
	}
	for field.Kind() == reflect.Interface || field.Kind() == reflect.Pointer {
		if field.IsNil() {
			return 0, false
		}
		field = field.Elem()
	}
	switch field.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return field.Uint(), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value := field.Int()
		if value < 0 {
			return 0, false
		}
		return uint64(value), true
	default:
		return 0, false
	}
}

func stringField(v reflect.Value, name string) (string, bool) {
	field := v.FieldByName(name)
	if !field.IsValid() {
		return "", false
	}
	for field.Kind() == reflect.Interface || field.Kind() == reflect.Pointer {
		if field.IsNil() {
			return "", false
		}
		field = field.Elem()
	}
	if field.Kind() != reflect.String {
		return "", false
	}
	return field.String(), true
}

func buildSourceSnippet(source string, pos sourcePosition) *SourceSnippet {
	lines := strings.Split(source, "\n")
	lineIndex := int(pos.Line) - 1
	if lineIndex < 0 || lineIndex >= len(lines) {
		return nil
	}

	text := strings.TrimSuffix(lines[lineIndex], "\r")
	column := int(pos.Column)
	if column < 1 {
		column = 1
	}
	runes := []rune(text)
	markerColumn := column - 1
	if markerColumn > len(runes) {
		markerColumn = len(runes)
	}

	return &SourceSnippet{
		FileName: pos.FileName,
		Line:     pos.Line,
		Column:   pos.Column,
		Text:     text,
		Marker:   strings.Repeat(" ", markerColumn) + "^",
	}
}

func scalarString(v reflect.Value, fallback string) string {
	if v.IsValid() && v.CanInterface() {
		if stringer, ok := v.Interface().(fmt.Stringer); ok {
			return stringer.String()
		}
	}
	return fallback
}

func quoteString(value string) string {
	const max = 140
	value = strings.ReplaceAll(value, "\n", `\n`)
	if len([]rune(value)) > max {
		runes := []rune(value)
		value = string(runes[:max]) + "..."
	}
	return strconv.Quote(value)
}

func (self *treeBuilder) id() string {
	self.nextID++
	return "n" + strconv.Itoa(self.nextID)
}

func valueInterface(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}
	if v.CanInterface() {
		return v.Interface()
	}
	return "<unexported>"
}
