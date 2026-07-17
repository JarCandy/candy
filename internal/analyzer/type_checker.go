package analyzer

import (
	"strings"

	candyerrors "github.com/CandyCrafts/candy/internal/errors"
	"github.com/CandyCrafts/candy/internal/parser"
	"github.com/CandyCrafts/candy/internal/parser/token"
	diagnostics "github.com/rp1s/digreyt"
)

type TypeChecker struct {
	Diagnostics *diagnostics.Arena
}

type TypeCheckResult struct {
	Checked bool
	Error   *TypeError
}

type TypeError struct {
	Name     string
	Declared string
	Got      string
	Pos      token.Position
}

type typePath []string

func NewTypeChecker(diagnostics *diagnostics.Arena) *TypeChecker {
	return &TypeChecker{Diagnostics: diagnostics}
}

func (self *TypeChecker) CheckLet(source []byte, let *parser.Let) TypeCheckResult {
	if let == nil {
		return TypeCheckResult{}
	}

	declared := typeFromAST(source, let.Type)
	got := exprType(source, let.Defualt)
	if declared.empty() || got.empty() {
		return TypeCheckResult{}
	}
	if declared.equal(got) {
		return TypeCheckResult{Checked: true}
	}

	declaredText := declared.String()
	gotText := got.String()
	typeErr := TypeError{
		Name:     tokenText(source, let.Name),
		Declared: declaredText,
		Got:      gotText,
		Pos:      let.Name.Pos,
	}

	if self.Diagnostics != nil {
		self.Diagnostics.Add(candyerrors.AnalyzerTypeMismatch(span(let.Name), declaredText, gotText))
	}

	return TypeCheckResult{Checked: true, Error: &typeErr}
}

func typeFromAST(source []byte, typ parser.Type) typePath {
	if typ == nil {
		return nil
	}

	switch n := typ.(type) {
	case *parser.TypeExpr:
		return tokenPath(source, n.Path)
	case parser.TypeExpr:
		return tokenPath(source, n.Path)
	default:
		return nil
	}
}

func exprType(source []byte, expr *parser.Expr) typePath {
	if expr == nil || *expr == nil {
		return nil
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
		return binaryExprType(source, n.Left, n.Right)
	case *parser.BinaryExpr:
		return binaryExprType(source, n.Left, n.Right)
	default:
		return nil
	}
}

func binaryExprType(source []byte, leftExpr *parser.Expr, rightExpr *parser.Expr) typePath {
	left := exprType(source, leftExpr)
	right := exprType(source, rightExpr)
	if !left.empty() && left.equal(right) {
		return left
	}
	return nil
}

func literalType(kind token.Kind) typePath {
	switch kind {
	case token.STRING, token.RAW_STRING:
		return typePath{"string"}
	case token.CHARACTER:
		return typePath{"char"}
	case token.INTEGER:
		return typePath{"int"}
	case token.FLOATING:
		return typePath{"float"}
	case token.IMAGINARY:
		return typePath{"complex"}
	case token.TRUE, token.FALSE:
		return typePath{"bool"}
	default:
		return nil
	}
}

func tokenPath(source []byte, path []token.Token) typePath {
	if len(path) == 0 {
		return nil
	}

	parts := make(typePath, 0, len(path))
	for _, tk := range path {
		parts = append(parts, tokenText(source, tk))
	}
	return parts
}

func (self typePath) empty() bool {
	return len(self) == 0
}

func (self typePath) equal(other typePath) bool {
	if len(self) != len(other) {
		return false
	}
	for i := range self {
		if self[i] != other[i] {
			return false
		}
	}
	return true
}

func (self typePath) String() string {
	return strings.Join(self, "::")
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
