package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/caramelang/caramel/pkg/branding"
)

const pluginDownloadTimeout = 5 * time.Minute

// DownloadPluginURL downloads a Git repository and installs it into
// ~/ProjectName/plugins/<repository-name>. Version may be a branch or tag.
// It returns the absolute path to the installed plugin directory.
func DownloadPluginURL(repositoryURL string, version string) (string, error) {
	repositoryURL = strings.TrimSpace(repositoryURL)
	version = strings.TrimSpace(version)
	if repositoryURL == "" {
		return "", fmt.Errorf("download plugin URL: repository URL is empty")
	}
	if version == "" {
		return "", fmt.Errorf("download plugin URL: version is empty")
	}
	if strings.HasPrefix(repositoryURL, "-") {
		return "", fmt.Errorf("download plugin URL: invalid repository URL %q", repositoryURL)
	}
	if strings.HasPrefix(version, "-") {
		return "", fmt.Errorf("download plugin URL: invalid version %q", version)
	}

	pluginName, err := repositoryName(repositoryURL)
	if err != nil {
		return "", fmt.Errorf("download plugin URL: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("download plugin URL: resolve home directory: %w", err)
	}
	pluginsDirectory := filepath.Join(home, branding.ProjectName, "plugins")
	if err := os.MkdirAll(pluginsDirectory, 0o755); err != nil {
		return "", fmt.Errorf("download plugin URL: create plugins directory: %w", err)
	}

	path, err := filepath.Abs(filepath.Join(pluginsDirectory, pluginName))
	if err != nil {
		return "", fmt.Errorf("download plugin: resolve path path: %w", err)
	}
	if _, err := os.Lstat(path); err == nil {
		return "", fmt.Errorf("download plugin URL: plugin %q is already installed at %q", pluginName, path)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("download plugin URL: inspect path %q: %w", path, err)
	}

	temporaryDirectory, err := os.MkdirTemp(
		"",
		fmt.Sprintf("%s-plugin-*", branding.ProjectName),
	)
	if err != nil {
		return "", fmt.Errorf("download plugin URL: create temporary directory: %w", err)
	}
	defer os.RemoveAll(temporaryDirectory)

	checkoutDirectory := filepath.Join(temporaryDirectory, "checkout")
	ctx, cancel := context.WithTimeout(context.Background(), pluginDownloadTimeout)
	defer cancel()
	if err := cloneRepository(ctx, repositoryURL, version, checkoutDirectory); err != nil {
		return "", err
	}

	stagingDirectory, err := os.MkdirTemp(pluginsDirectory, "."+pluginName+"-*")
	if err != nil {
		return "", fmt.Errorf("download plugin URL: create staging directory: %w", err)
	}
	removeStaging := true
	defer func() {
		if removeStaging {
			_ = os.RemoveAll(stagingDirectory)
		}
	}()

	if err := copyRepository(checkoutDirectory, stagingDirectory); err != nil {
		return "", fmt.Errorf("download plugin URL: prepare plugin files: %w", err)
	}
	if err := os.Rename(stagingDirectory, path); err != nil {
		return "", fmt.Errorf("download plugin URL: install plugin at %q: %w", path, err)
	}
	removeStaging = false

	return path, nil
}

// DownloadLatestPluginURL installs the latest published repository release.
func DownloadLatestPluginURL(repositoryURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), pluginDownloadTimeout)
	defer cancel()

	version, err := latestReleaseVersion(ctx, repositoryURL)
	if err != nil {
		return "", err
	}
	return DownloadPluginURL(repositoryURL, version)
}

func latestReleaseVersion(ctx context.Context, repositoryURL string) (string, error) {
	parsed, err := parseRepositoryURL(repositoryURL)
	if err != nil {
		return "", fmt.Errorf("download latest plugin: %w", err)
	}

	var endpoint string
	switch {
	case parsed.host == "github.com":
		endpoint = fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", parsed.path)
	case parsed.host == "gitlab.com" || strings.Contains(parsed.host, "gitlab"):
		endpoint = fmt.Sprintf("%s://%s/api/v4/projects/%s/releases/permalink/latest", parsed.scheme, parsed.host, url.PathEscape(parsed.path))
	default:
		// Gitea and Forgejo expose the same latest-release endpoint.
		endpoint = fmt.Sprintf("%s://%s/api/v1/repos/%s/releases/latest", parsed.scheme, parsed.host, parsed.path)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("download latest plugin: create release request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "caramel-plugin-downloader")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("download latest plugin: request latest release: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download latest plugin: release API returned %s", response.Status)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&release); err != nil {
		return "", fmt.Errorf("download latest plugin: decode release: %w", err)
	}
	release.TagName = strings.TrimSpace(release.TagName)
	if release.TagName == "" {
		return "", fmt.Errorf("download latest plugin: release has no tag name")
	}
	return release.TagName, nil
}

type repositoryLocation struct {
	scheme string
	host   string
	path   string
}

func parseRepositoryURL(repositoryURL string) (repositoryLocation, error) {
	repositoryURL = strings.TrimSpace(repositoryURL)
	parsed, err := url.Parse(repositoryURL)
	if err != nil || parsed.Host == "" {
		return repositoryLocation{}, fmt.Errorf("release lookup requires an HTTP(S) repository URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return repositoryLocation{}, fmt.Errorf("release lookup does not support URL scheme %q", parsed.Scheme)
	}
	repositoryPath := strings.TrimSuffix(strings.Trim(parsed.Path, "/"), ".git")
	if repositoryPath == "" || !strings.Contains(repositoryPath, "/") {
		return repositoryLocation{}, fmt.Errorf("invalid repository URL %q", repositoryURL)
	}
	return repositoryLocation{
		scheme: parsed.Scheme,
		host:   strings.ToLower(parsed.Host),
		path:   repositoryPath,
	}, nil
}

func cloneRepository(ctx context.Context, repositoryURL, version, path string) error {
	args := []string{"clone", "--depth", "1"}
	if version != "" {
		args = append(args, "--branch", version)
	}
	args = append(args, "--", repositoryURL, path)

	command := exec.CommandContext(ctx, "git", args...)
	command.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	output, err := command.CombinedOutput()
	if err == nil {
		return nil
	}
	if ctx.Err() != nil {
		return fmt.Errorf("download plugin URL: clone timed out after %s: %w", pluginDownloadTimeout, ctx.Err())
	}
	message := strings.TrimSpace(string(output))
	if message == "" {
		message = err.Error()
	}
	return fmt.Errorf("download plugin URL: git clone failed: %s", message)
}

func repositoryName(repositoryURL string) (string, error) {
	path := repositoryURL
	if parsed, err := url.Parse(repositoryURL); err == nil && parsed.Scheme != "" {
		path = parsed.Path
	} else if colon := strings.LastIndex(repositoryURL, ":"); colon >= 0 {
		path = repositoryURL[colon+1:]
	}

	path = strings.TrimSuffix(strings.TrimRight(path, "/"), ".git")
	name := filepath.Base(path)
	if name == "" || name == "." || name == string(filepath.Separator) || !filepath.IsLocal(name) {
		return "", fmt.Errorf("cannot determine repository name from %q", repositoryURL)
	}
	return name, nil
}

func copyRepository(source, destination string) error {
	return filepath.WalkDir(source, func(sourcePath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relativePath, err := filepath.Rel(source, sourcePath)
		if err != nil {
			return err
		}
		if relativePath == ".git" || strings.HasPrefix(relativePath, ".git"+string(filepath.Separator)) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if relativePath == "." {
			return nil
		}

		target := filepath.Join(destination, relativePath)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symbolic links are not supported: %s", relativePath)
		}
		if entry.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported file type: %s", relativePath)
		}
		return copyFile(sourcePath, target, info.Mode().Perm())
	})
}

func copyFile(source, path string, mode os.FileMode) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(output, input); err != nil {
		_ = output.Close()
		return err
	}
	return output.Close()
}
