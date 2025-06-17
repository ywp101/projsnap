package apps

import (
	"projctx/utils"
)

type PackConfig []string

func NewAppConfigs(appName string) []AppConfig {
	return []AppConfig{{
		AppName: appName,
		Args:    make(PackConfig, 0),
	}}
}

func NewAppConfigsWithArgs(appName string, args PackConfig) []AppConfig {
	return []AppConfig{{
		AppName: appName,
		Args:    args,
	}}
}

type AppConfig struct {
	AppName string     `json:"app_name"`
	Args    PackConfig `json:"args"`
}

type AppPacker interface {
	Pack(configDir string, appName string) ([]AppConfig, error)
	Unpack(*AppConfig, bool) error
	Quit(string) error
}

type NormalPacker struct {
}

func (NormalPacker) Pack(_, appName string) ([]AppConfig, error) {
	return NewAppConfigs(appName), nil
}

func (NormalPacker) Unpack(ws *AppConfig, running bool) error {
	if !running {
		return utils.OpenApp(ws.AppName)
	}
	return nil
}

func (NormalPacker) Quit(appName string) error {
	return utils.GracefulQuit(appName)
}
