package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	cacheclitext "github.com/CandyCrafts/candy/internal/database/cache/cache_cli_text"
	cacheclitextsql "github.com/CandyCrafts/candy/internal/database/cache/cache_cli_text/sql"
	"github.com/CandyCrafts/candy/pkg/branding"

	_ "github.com/mattn/go-sqlite3"
)

type CLITextEntry = cacheclitext.CacheCliText

type SQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type CacheDatabase interface {
	Init(ctx context.Context) error
	CLIText() CLITextCache
}

type CLITextCache interface {
	Init(ctx context.Context) error
	Create(ctx context.Context, entry CLITextEntry) error
	Get(ctx context.Context, filters map[string]any) (CLITextEntry, error)
	Update(ctx context.Context, entry CLITextEntry, filters map[string]any) error
	Delete(ctx context.Context, filters map[string]any) error
	List(ctx context.Context, limit int, offset int) ([]CLITextEntry, error)
	Search(ctx context.Context, term string, limit int, offset int) ([]CLITextEntry, error)
	ReadText(ctx context.Context, originalText string, lang string) (string, bool, error)
	WriteText(ctx context.Context, originalText string, lang string, text string) error
}

type sqlCacheDatabase struct {
	cliText CLITextCache
}

var _ CacheDatabase = (*sqlCacheDatabase)(nil)

func NewCacheDatabase(executor SQLExecutor) CacheDatabase {
	return &sqlCacheDatabase{
		cliText: NewCLITextCache(executor),
	}
}

func (self *sqlCacheDatabase) Init(ctx context.Context) error {
	return self.cliText.Init(ctx)
}

func (self *sqlCacheDatabase) CLIText() CLITextCache {
	return self.cliText
}

type sqlCLITextCache struct {
	executor SQLExecutor
}

var _ CLITextCache = (*sqlCLITextCache)(nil)

func NewCLITextCache(executor SQLExecutor) CLITextCache {
	return &sqlCLITextCache{executor: executor}
}

func (self *sqlCLITextCache) Init(ctx context.Context) error {
	return cacheclitextsql.Init(ctx, self.executor)
}

func (self *sqlCLITextCache) Create(ctx context.Context, entry CLITextEntry) error {
	return cacheclitextsql.Create(ctx, self.executor, entry)
}

func (self *sqlCLITextCache) Get(ctx context.Context, filters map[string]any) (CLITextEntry, error) {
	return cacheclitextsql.Get(ctx, self.executor, filters)
}

func (self *sqlCLITextCache) Update(ctx context.Context, entry CLITextEntry, filters map[string]any) error {
	return cacheclitextsql.Update(ctx, self.executor, entry, filters)
}

func (self *sqlCLITextCache) Delete(ctx context.Context, filters map[string]any) error {
	return cacheclitextsql.Delete(ctx, self.executor, filters)
}

func (self *sqlCLITextCache) List(ctx context.Context, limit int, offset int) ([]CLITextEntry, error) {
	return cacheclitextsql.List(ctx, self.executor, limit, offset)
}

func (self *sqlCLITextCache) Search(ctx context.Context, term string, limit int, offset int) ([]CLITextEntry, error) {
	return cacheclitextsql.Search(ctx, self.executor, term, limit, offset)
}

func (self *sqlCLITextCache) ReadText(ctx context.Context, originalText string, lang string) (string, bool, error) {
	model, err := self.Get(ctx, map[string]any{"OriginalText": originalText, "Lang": lang})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}

		return "", false, err
	}

	return model.Text, true, nil
}

func (self *sqlCLITextCache) WriteText(ctx context.Context, originalText string, lang string, text string) error {
	entry := CLITextEntry{
		Lang:         lang,
		OriginalText: originalText,
		Text:         text,
	}

	_, ok, err := self.ReadText(ctx, originalText, lang)
	if err != nil {
		return err
	}
	if !ok {
		return self.Create(ctx, entry)
	}

	return self.Update(ctx, entry, map[string]any{"OriginalText": originalText, "Lang": lang})
}

func OpenDatabase(name string) (*sql.DB, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("get user config directory: %w", err)
	}

	appDir := filepath.Join(configDir, branding.ProjectName)

	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, fmt.Errorf("create app directory: %w", err)
	}

	path := filepath.Join(appDir, name)
	fmt.Println(path)

	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	return conn, nil
}
