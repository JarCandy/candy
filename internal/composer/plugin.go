package composer

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/CandyCrafts/candy/pkg/branding"
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
	self.Custom.InterfacePath = valueOrDefault(self.Custom.InterfacePath, "interface"+branding.PrefixInterfaceFile)
}

func DownloadProjectPlagin(url string, localPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	outFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}
