package composer

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/caramelang/caramel/pkg/branding"
)

type TypePlugin uint8

const (
	Base TypePlugin = iota
)

type Custom struct {
	PluginPath      *string
	DescriptionPath *string
	InterfacePath   *string
}

type ProjectPlugin struct {
	Type    *TypePlugin
	Name    *string
	Version *string
	Author  *string

	Custom Custom
}

func valueOrDefault[T any](value *T, defaultValue T) *T {
	if value != nil {
		return value
	}

	return &defaultValue
}

func (self *ProjectPlugin) Default(projectPatch string) {
	self.Type = valueOrDefault(self.Type, Base)
	self.Name = valueOrDefault(self.Name, "default-plugin")
	self.Version = valueOrDefault(self.Version, "latest")
	self.Author = valueOrDefault(self.Author, "pidoras")

	self.Custom.PluginPath = valueOrDefault(self.Custom.PluginPath, "out/plugin.wasm")
	self.Custom.DescriptionPath = valueOrDefault(self.Custom.DescriptionPath, "d.txt")
	self.Custom.InterfacePath = valueOrDefault(self.Custom.InterfacePath, "interface"+branding.InterfaceFileExtension)
}

func DownloadProjectPlugin(url string, localPath string) error {
	return downloadProjectPlugin(http.DefaultClient, url, localPath)
}

// DownloadProjectPlagin is kept for compatibility with the old misspelled API.
func DownloadProjectPlagin(url string, localPath string) error {
	return DownloadProjectPlugin(url, localPath)
}

func downloadProjectPlugin(client *http.Client, url string, localPath string) (resultErr error) {
	if client == nil {
		return errors.New("http client is nil")
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("close response body: %w", err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	outFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("close output file: %w", err))
		}
	}()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}
