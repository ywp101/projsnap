package apps

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"projctx/utils"
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

func (o Obsidian) Pack(configDir, _ string) ([]string, error) {
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

func (o Obsidian) Unpack(ws *AppConfig, running bool) error {
	if running {
		_ = o.Quit(ws.AppName)
	}
	if err := utils.RecoverBakFile(ws.Args[0]); err != nil {
		return err
	}
	return utils.OpenApp(ws.AppName)
}

func (o Obsidian) Quit(appName string) error {
	return utils.GracefulQuit(appName)
}
