package analyzer

import (
	stderrors "errors"

	"github.com/CandyCrafts/candy/internal/composer"
	"github.com/CandyCrafts/candy/internal/parser"
	diagnostics "github.com/rp1s/digreyt"
)

type Analyzer struct {
	Diagnostics *diagnostics.Arena
	Types       *TypeChecker
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

func New() *Analyzer {
	arena := diagnostics.New("")
	return &Analyzer{Diagnostics: arena, Types: NewTypeChecker(arena)}
}

func (a *Analyzer) Project(project *composer.Project) (*Result, error) {
	if project == nil {
		return nil, stderrors.New("project is nil")
	}
	if a.Diagnostics == nil {
		a.Diagnostics = diagnostics.New("")
	}
	if a.Types == nil {
		a.Types = NewTypeChecker(a.Diagnostics)
	} else {
		a.Types.Diagnostics = a.Diagnostics
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

	result := a.Types.CheckLet(source, let)
	if result.Checked {
		file.TypeChecks++
	}
	if result.Error != nil {
		file.TypeErrors = append(file.TypeErrors, *result.Error)
	}
}
