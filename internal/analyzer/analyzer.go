package analyzer

import (
	stderrors "errors"
	"strings"

	"github.com/CandyCrafts/candy/internal/composer"
	candyerrors "github.com/CandyCrafts/candy/internal/errors"
	"github.com/CandyCrafts/candy/internal/parser"
	"github.com/CandyCrafts/candy/internal/parser/token"
	diagnostics "github.com/rp1s/digreyt"
)

type Analyzer struct {
	Diagnostics *diagnostics.Arena
}

type Result struct {
	ProjectName string
	Files       []File
	Diagnostics *diagnostics.Arena
}

type File struct {
	Name string
	Path string

	LetCount   int
	TypeChecks int
	TypeErrors []TypeError
}

type TypeError struct {
	Name     string
	Declared string
	Got      string
	Pos      token.Position
}

func New() *Analyzer {
	return &Analyzer{
		Diagnostics: diagnostics.New(""),
	}
}

func (a *Analyzer) Project(project *composer.Project) (*Result, error) {
	if project == nil {
		return nil, stderrors.New("project is nil")
	}
	if a.Diagnostics == nil {
		a.Diagnostics = diagnostics.New("")
	}

	result := &Result{
		ProjectName: project.Name,
		Files:       make([]File, 0, len(project.AstFile)),
		Diagnostics: a.Diagnostics,
	}

	for _, astFile := range project.AstFile {
		if a.Diagnostics.Source == "" {
			a.Diagnostics.Source = string(astFile.Source)
		}

		file := File{
			Name: astFile.FileName,
			Path: astFile.Path,
		}
		a.checkAST(&file, astFile.Source, astFile.Ast)
		result.Files = append(result.Files, file)
	}

	return result, nil
}

func (a *Analyzer) checkAST(file *File, source []byte, ast parser.AST) {
	for _, decl := range ast.Decls {
		a.checkDecl(file, source, decl)
	}
}

func (a *Analyzer) checkDecl(file *File, source []byte, decl parser.Decl) {
	switch n := decl.(type) {
	case *parser.LetDecl:
		a.checkLet(file, source, n.Let)
	case *parser.QualifiedDecl:
		for _, stmt := range n.Body {
			a.checkStmt(file, source, stmt)
		}
	}
}

func (a *Analyzer) checkStmt(file *File, source []byte, stmt parser.Stmt) {
	if n, ok := stmt.(*parser.LetStmt); ok {
		a.checkLet(file, source, n.Let)
	}
}

func (a *Analyzer) checkLet(file *File, source []byte, let *parser.Let) {
	if let == nil {
		return
	}
	file.LetCount++

	declared := typeText(source, let.Type)
	got := exprType(source, let.Defualt)
	if declared == "" || got == "" {
		return
	}

	file.TypeChecks++
	if declared == got {
		return
	}

	typeErr := TypeError{
		Name:     tokenText(source, let.Name),
		Declared: declared,
		Got:      got,
		Pos:      let.Name.Pos,
	}
	file.TypeErrors = append(file.TypeErrors, typeErr)
	a.Diagnostics.Add(candyerrors.AnalyzerTypeMismatch(span(let.Name), declared, got))
}

func typeText(source []byte, typ parser.Type) string {
	if typ == nil {
		return ""
	}

	switch n := typ.(type) {
	case *parser.TypeExpr:
		return tokenPath(source, n.Path)
	case parser.TypeExpr:
		return tokenPath(source, n.Path)
	default:
		return ""
	}
}

func exprType(source []byte, expr *parser.Expr) string {
	if expr == nil || *expr == nil {
		return ""
	}

	switch n := (*expr).(type) {
	case parser.LiteralExpr:
		return literalType(n.Value.Kind)
	case *parser.LiteralExpr:
		return literalType(n.Value.Kind)
	case parser.UnaryExpr:
		return exprType(source, n.X)
	case *parser.UnaryExpr:
		return exprType(source, n.X)
	case parser.BinaryExpr:
		left := exprType(source, n.Left)
		right := exprType(source, n.Right)
		if left != "" && left == right {
			return left
		}
		return ""
	case *parser.BinaryExpr:
		left := exprType(source, n.Left)
		right := exprType(source, n.Right)
		if left != "" && left == right {
			return left
		}
		return ""
	default:
		return ""
	}
}

func literalType(kind token.Kind) string {
	switch kind {
	case token.STRING, token.RAW_STRING:
		return "string"
	case token.CHARACTER:
		return "char"
	case token.INTEGER:
		return "int"
	case token.FLOATING:
		return "float"
	case token.IMAGINARY:
		return "complex"
	case token.TRUE, token.FALSE:
		return "bool"
	default:
		return ""
	}
}

func tokenPath(source []byte, path []token.Token) string {
	if len(path) == 0 {
		return ""
	}

	parts := make([]string, 0, len(path))
	for _, tk := range path {
		parts = append(parts, tokenText(source, tk))
	}
	return strings.Join(parts, "::")
}

func tokenText(source []byte, tk token.Token) string {
	if len(source) == 0 || tk.End > uint64(len(source)) || tk.Start > tk.End {
		return ""
	}
	return string(tk.Literal(&source))
}

func span(tk token.Token) candyerrors.Span {
	return candyerrors.Span{
		Start: tk.Start,
		End:   tk.End,
		Pos: candyerrors.Position{
			FileName: tk.Pos.FileName,
			Line:     tk.Pos.Line,
			Column:   tk.Pos.Column,
			Offset:   tk.Pos.Offset,
		},
	}
}
