package apps

import (
	"projctx/utils"
)

type AppConfig struct {
	AppName string   `json:"app_name"`
	Args    []string `json:"args"`
}

type AppPacker interface {
	Pack(string, string) ([]string, error)
	Unpack(*AppConfig, bool) error
	Quit(string) error
}

type NormalPacker struct {
}

func (NormalPacker) Pack(_, _ string) ([]string, error) {
	return []string{}, nil
}

func (NormalPacker) Unpack(ws *AppConfig, running bool) error {
	if !running {
		return utils.OpenApp(ws.AppName, ws.Args...)
	}
	return nil
}

func (NormalPacker) Quit(appName string) error {
	return utils.GracefulQuit(appName)
}
