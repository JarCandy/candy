# lipa

`lipa` is a tiny independent Go library for visualizing Go values as a clean
interactive tree in a browser window.

It uses reflection, follows pointers and interfaces into real objects, protects
against cycles, and renders a draggable/zoomable tree.

## Usage

```go
package main

import "github.com/rp1s/lipa"

func main() {
	type User struct {
		Name string
		Tags []string
	}

	user := &User{Name: "Candy", Tags: []string{"parser", "ast"}}
	_ = lipa.View(user, lipa.WithTitle("Candy AST"))
}
```

Use `Render` if you only need HTML:

```go
html, err := lipa.Render(user, lipa.WithoutOpen())
```
