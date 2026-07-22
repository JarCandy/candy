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
	"github.com/caramelang/caramel/pkg/branding"
)

var wasmV1Header = []byte{
	0x00, 0x61, 0x73, 0x6d,
	0x01, 0x00, 0x00, 0x00,
}

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
	infoPath := filepath.Join(pluginDirectory, branding.PluginInfoFileName)
	info, err := os.Stat(infoPath)
	if err != nil {
		return pluginpubinfo.PluginPubInfo{}, fmt.Errorf("load plugin: inspect %q: %w", infoPath, err)
	}
	if !info.Mode().IsRegular() {
		return pluginpubinfo.PluginPubInfo{}, fmt.Errorf("load plugin: %q is not a regular file", infoPath)
	}

	content, err := os.ReadFile(infoPath)
	if err != nil {
		return pluginpubinfo.PluginPubInfo{}, fmt.Errorf("load plugin: read %q: %w", infoPath, err)
	}
	var pubInfo pluginpubinfo.PluginPubInfo
	if err := json.Unmarshal(content, &pubInfo); err != nil {
		return pluginpubinfo.PluginPubInfo{}, fmt.Errorf("load plugin: parse %q: %w", infoPath, err)
	}
	return pubInfo, nil
}

func readPluginWasm(pluginDirectory string) (string, []byte, error) {
	candidates := []string{
		filepath.Join(pluginDirectory, branding.PluginWasmFileName),
		filepath.Join(pluginDirectory, branding.PluginOutputDirectory, branding.PluginWasmFileName),
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
		if len(wasmBin) < len(wasmV1Header) || !bytes.Equal(wasmBin[:len(wasmV1Header)], wasmV1Header[:]) {
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
		filepath.Join(pluginDirectory, branding.PluginOutputDirectory),
	)
}
