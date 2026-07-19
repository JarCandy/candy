package analyzer

import (
	stderrors "errors"

	"github.com/caramelang/caramel/internal/composer"
	"github.com/caramelang/caramel/internal/parser"
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

func (self *Analyzer) Project(project *composer.Project) (*Result, error) {
	if project == nil {
		return nil, stderrors.New("project is nil")
	}
	if self.Diagnostics == nil {
		self.Diagnostics = diagnostics.New("")
	}
	if self.Types == nil {
		self.Types = NewTypeChecker(self.Diagnostics)
	} else {
		self.Types.Diagnostics = self.Diagnostics
	}

	result := &Result{
		ProjectName: project.Name,
		Files:       make([]File, 0, len(project.AstFile)),
		Diagnostics: self.Diagnostics,
	}

	for _, astFile := range project.AstFile {
		if self.Diagnostics.Source == "" {
			self.Diagnostics.Source = string(astFile.Source)
		}

		file := File{
			Name: astFile.FileName,
			Path: astFile.Path,
		}
		self.checkAST(&file, astFile.Source, astFile.Ast)
		result.Files = append(result.Files, file)
	}

	return result, nil
}

func (self *Analyzer) checkAST(file *File, source []byte, ast parser.AST) {
	for _, decl := range ast.Decls {
		self.checkDecl(file, source, decl)
	}
}

func (self *Analyzer) checkDecl(file *File, source []byte, decl parser.Decl) {
	switch n := decl.(type) {
	case *parser.LetDecl:
		self.checkLet(file, source, n.Let)
	case *parser.QualifiedDecl:
		for _, stmt := range n.Body {
			self.checkStmt(file, source, stmt)
		}
	}
}

func (self *Analyzer) checkStmt(file *File, source []byte, stmt parser.Stmt) {
	if n, ok := stmt.(*parser.LetStmt); ok {
		self.checkLet(file, source, n.Let)
	}
}

func (self *Analyzer) checkLet(file *File, source []byte, let *parser.Let) {
	if let == nil {
		return
	}
	file.LetCount++

	result := self.Types.CheckLet(source, let)
	if result.Checked {
		file.TypeChecks++
	}
	if result.Error != nil {
		file.TypeErrors = append(file.TypeErrors, *result.Error)
	}
}
