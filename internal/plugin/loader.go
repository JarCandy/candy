package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/caramelang/caramel/internal/models/plugin"
	pluginpubinfo "github.com/caramelang/caramel/internal/models/plugin_pub_info"
)

var wasmHeader = []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

type LoadPluginArgs struct {
	Path    *string
	Url     *string
	Version *string
}

func LoadPlugin(l LoadPluginArgs) (*plugin.Plugin, error) {
	if (l.Path == nil) == (l.Url == nil) {
		return nil, fmt.Errorf("load plugin: specify exactly one of Path or Url")
	}

	var path string
	var isGlobal bool
	var err error
	if l.Path != nil {
		path = strings.TrimSpace(*l.Path)
		if path == "" {
			return nil, fmt.Errorf("load plugin: path is empty")
		}
		path, err = filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("load plugin: resolve path: %w", err)
		}
	}
	if l.Url != nil {
		isGlobal = true
		if l.Version != nil {
			path, err = DownloadPluginURL(*l.Url, *l.Version)
		} else {
			path, err = DownloadLatestPluginURL(*l.Url)
		}
		if err != nil {
			return nil, err
		}
	}

	pubInfo, err := readPluginPubInfo(path)
	if err != nil {
		return nil, err
	}

	_, wasmBin, err := readPluginWasm(path)
	if err != nil {
		return nil, err
	}

	return &plugin.Plugin{
		PubInfo:        pubInfo,
		IsGlobal:       isGlobal,
		PathPluginSave: path,
		WasmBin:        wasmBin,
	}, nil
}

func readPluginPubInfo(pluginDirectory string) (pluginpubinfo.PluginPubInfo, error) {
	interfacePath := filepath.Join(pluginDirectory, "interface.json")
	info, err := os.Stat(interfacePath)
	if err != nil {
		return pluginpubinfo.PluginPubInfo{}, fmt.Errorf("load plugin: inspect %q: %w", interfacePath, err)
	}
	if !info.Mode().IsRegular() {
		return pluginpubinfo.PluginPubInfo{}, fmt.Errorf("load plugin: %q is not a regular file", interfacePath)
	}

	content, err := os.ReadFile(interfacePath)
	if err != nil {
		return pluginpubinfo.PluginPubInfo{}, fmt.Errorf("load plugin: read %q: %w", interfacePath, err)
	}
	var pubInfo pluginpubinfo.PluginPubInfo
	if err := json.Unmarshal(content, &pubInfo); err != nil {
		return pluginpubinfo.PluginPubInfo{}, fmt.Errorf("load plugin: parse %q: %w", interfacePath, err)
	}
	return pubInfo, nil
}

func readPluginWasm(pluginDirectory string) (string, []byte, error) {
	candidates := []string{
		filepath.Join(pluginDirectory, "plugin.wasm"),
		filepath.Join(pluginDirectory, "out", "plugin.wasm"),
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", nil, fmt.Errorf("load plugin: inspect %q: %w", candidate, err)
		}
		if !info.Mode().IsRegular() {
			return "", nil, fmt.Errorf("load plugin: %q is not a regular file", candidate)
		}

		wasmBin, err := os.ReadFile(candidate)
		if err != nil {
			return "", nil, fmt.Errorf("load plugin: read %q: %w", candidate, err)
		}
		if len(wasmBin) < len(wasmHeader) || !bytes.Equal(wasmBin[:len(wasmHeader)], wasmHeader) {
			return "", nil, fmt.Errorf("load plugin: %q is not a valid WebAssembly module", candidate)
		}

		wasmPath, err := filepath.Abs(candidate)
		if err != nil {
			return "", nil, fmt.Errorf("load plugin: resolve WebAssembly path: %w", err)
		}
		return wasmPath, wasmBin, nil
	}

	return "", nil, fmt.Errorf(
		"load plugin: plugin.wasm not found in %q or %q",
		pluginDirectory,
		filepath.Join(pluginDirectory, "out"),
	)
}
