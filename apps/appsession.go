package apps

import (
	"workspace/utils"
)

type WorkspaceConfig struct {
	AppName string   `json:"app_name"`
	Args    []string `json:"args"`
}

type AppSessionPacker interface {
	Pack(string, string) ([]string, error)
	Unpack(*WorkspaceConfig) error
	Quit(string) error
}

type NormalPacker struct {
}

func (NormalPacker) Pack(_, _ string) ([]string, error) {
	return []string{}, nil
}

func (NormalPacker) Unpack(ws *WorkspaceConfig) error {
	return utils.OpenApp(ws.AppName, ws.Args...)
}

func (NormalPacker) Quit(appName string) error {
	return utils.GracefulQuit(appName)
}
