package apps

import (
	"workspace/utils"
)

type JetBrains struct {
}

func (f JetBrains) Pack(ideName string) ([]string, error) {
	return utils.RunOsascript(ideName)
}

func (f JetBrains) Unpack(ws *WorkspaceConfig) error {
	return utils.OpenApp(ws.AppName, ws.Args...)
}
