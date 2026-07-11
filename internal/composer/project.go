package composer

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/CandyCrafts/candy/internal/parser"
)

type Project struct {
	WorkPath string // home - project
	AstFile  []AstFile
}

type AstFile struct {
	FileName string
	Ast      parser.AST
}

func Load(fileName string, dir string) (*Project, error) {
	path, err := home_path()
	if err != nil {
		return nil, err
	}

	content, err := get_content_file(fileName, dir)
	if err != nil {
		return nil, err
	}
	// TODO
	_ = content

	p := &Project{
		WorkPath: path,
	}
	return p, nil
}

func home_path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	path, err := filepath.Rel(home, cwd)
	if err != nil {
		return "", err
	}

	return path, nil
}

func get_content_file(fileName string, dir string) ([]byte, error) {
	if dir == "" {
		dir = "./"
	}

	filePath := filepath.Join(dir, fileName)

	info, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}

		return nil, err
	}

	if info.IsDir() {
		return nil, fmt.Errorf("The specified path contains a directory, not a file")
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return content, nil
}
