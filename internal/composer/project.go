package composer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/caramelang/caramel/internal/parser"
)

type Project struct {
	Name     string
	WorkPath string // directory where the build command was called
	AstFile  []AstFile
}

type AstFile struct {
	FileName string
	Path     string
	Source   []byte
	Ast      parser.AST
}

func Load(filePath string, name string) (*Project, error) {
	if strings.TrimSpace(filePath) == "" {
		return nil, fmt.Errorf("build file path is required")
	}
	name = strings.TrimSpace(name)

	workPath, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	absPath, err := resolveFilePath(workPath, filePath)
	if err != nil {
		return nil, err
	}

	content, err := getContentFile(absPath)
	if err != nil {
		return nil, err
	}

	ast, err := parser.New(content, absPath).Run()
	if err != nil {
		return nil, err
	}

	if name == "" {
		name = nameFromPath(absPath)
	}

	return &Project{
		Name:     name,
		WorkPath: workPath,
		AstFile: []AstFile{
			{
				FileName: filepath.Base(absPath),
				Path:     absPath,
				Source:   content,
				Ast:      *ast,
			},
		},
	}, nil
}

func resolveFilePath(workPath string, filePath string) (string, error) {
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(workPath, filePath)
	}

	return filepath.Abs(filePath)
}

func getContentFile(filePath string) ([]byte, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, fmt.Errorf("the specified path contains a directory, not a file: %s", filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func nameFromPath(filePath string) string {
	name := filepath.Base(filePath)
	ext := filepath.Ext(name)
	if ext == "" {
		return name
	}
	return strings.TrimSuffix(name, ext)
}
