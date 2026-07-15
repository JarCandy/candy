package lipa

import (
	"strings"
	"testing"
)

type testNode struct {
	Name string
	Next *testNode
}

type testPosition struct {
	FileName string
	Line     uint64
	Column   uint64
	Offset   uint64
}

type testToken struct {
	Pos testPosition
}

type testLocatedNode struct {
	Tok  testToken
	Name string
}

type testSourceView struct {
	Source string
	Child  testLocatedNode
}

func TestBuildDereferencesPointersAndDetectsCycles(t *testing.T) {
	root := &testNode{Name: "root"}
	root.Next = root

	tree := Build(root)

	if tree.Kind != "struct" {
		t.Fatalf("expected root object kind, got %q", tree.Kind)
	}
	if len(tree.Children) != 2 {
		t.Fatalf("expected dereferenced fields, got %d", len(tree.Children))
	}
	if !containsCycle(tree) {
		t.Fatal("expected cycle node")
	}
}

func TestRenderIncludesInteractiveTree(t *testing.T) {
	html, err := Render(map[string]int{"one": 1}, WithoutOpen())
	if err != nil {
		t.Fatalf("Render() failed: %v", err)
	}
	for _, part := range []string{"Lipa Tree", "Search node", "Show hidden", "navigator", "selected", "const root ="} {
		if !strings.Contains(html, part) {
			t.Fatalf("expected html to contain %q", part)
		}
	}
}

func TestBuildMarksHiddenFields(t *testing.T) {
	tree := Build(testNode{Name: "root"}, WithHiddenFields("Next"))

	var next *Node
	for _, child := range tree.Children {
		if child.Name == "Next" {
			next = child
			break
		}
	}
	if next == nil {
		t.Fatal("expected Next child")
	}
	if !next.Hidden {
		t.Fatal("expected Next child to be hidden")
	}
}

func TestBuildAddsSourceSnippetFromToken(t *testing.T) {
	tree := Build(
		testLocatedNode{
			Tok: testToken{
				Pos: testPosition{
					FileName: "example.cm",
					Line:     2,
					Column:   3,
				},
			},
			Name: "node",
		},
		WithSource("first\n  second\nthird"),
		WithHiddenFields("Tok"),
	)

	if tree.Snippet == nil {
		t.Fatal("expected source snippet")
	}
	if tree.Snippet.FileName != "example.cm" || tree.Snippet.Line != 2 || tree.Snippet.Column != 3 {
		t.Fatalf("unexpected snippet position: %+v", tree.Snippet)
	}
	if tree.Snippet.Text != "  second" || tree.Snippet.Marker != "  ^" {
		t.Fatalf("unexpected snippet body: %+v", tree.Snippet)
	}
}

func TestBuildUsesRootSourceFieldForSnippets(t *testing.T) {
	tree := Build(testSourceView{
		Source: "first\n  second\nthird",
		Child: testLocatedNode{
			Tok: testToken{
				Pos: testPosition{
					FileName: "example.cm",
					Line:     2,
					Column:   3,
				},
			},
			Name: "node",
		},
	})

	child := childNamed(tree, "Child")
	if child == nil {
		t.Fatal("expected Child node")
	}
	if child.Snippet == nil {
		t.Fatal("expected source snippet from root Source field")
	}
	if child.Snippet.Text != "  second" || child.Snippet.Marker != "  ^" {
		t.Fatalf("unexpected snippet body: %+v", child.Snippet)
	}
}

func containsCycle(node *Node) bool {
	if node == nil {
		return false
	}
	if node.Cycle {
		return true
	}
	for _, child := range node.Children {
		if containsCycle(child) {
			return true
		}
	}
	return false
}

func childNamed(node *Node, name string) *Node {
	if node == nil {
		return nil
	}
	for _, child := range node.Children {
		if child.Name == name {
			return child
		}
	}
	return nil
}
