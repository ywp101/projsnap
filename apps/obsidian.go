package apps

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"workspace/utils"
)

var (
	obsidianConfigPath = "~/Library/Application Support/obsidian/obsidian.json"
)

type VaultsFile struct {
	Vaults map[string]Vault `json:"vaults"`
}

type Vault struct {
	Path string `json:"path"`
	Ts   int64  `json:"ts"`
	Open bool   `json:"open"`
}

type Obsidian struct {
}

func (Obsidian) Pack(configDir, _ string) ([]string, error) {
	configPath, _ := utils.ExpandUser(obsidianConfigPath)
	fd, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	valuts := VaultsFile{}
	if err := json.NewDecoder(fd).Decode(&valuts); err != nil {
		return nil, err
	}
	var workspacePath string
	for _, vault := range valuts.Vaults {
		if vault.Open {
			workspacePath = filepath.Join(vault.Path, ".obsidian", "workspace.json")
			break
		}
	}
	if workspacePath == "" {
		return nil, errors.New("no open vault")
	}
	if bakFile, err := utils.BakFile(configDir, workspacePath); err != nil {
		return nil, err
	} else {
		return []string{bakFile}, nil
	}
}

func (Obsidian) Unpack(ws *AppConfig) error {
	if err := utils.GracefulQuit(ws.AppName); err != nil {
		return nil
	}
	if err := utils.RecoverBakFile(ws.Args[0]); err != nil {
		return err
	}
	return utils.OpenApp(ws.AppName)
}

func (Obsidian) Quit(appName string) error {
	return utils.GracefulQuit(appName)
}
