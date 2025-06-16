package apps

import (
	"workspace/utils"
)

type Finder struct {
}

func (f Finder) Pack(_, _ string) ([]string, error) {
	script := `
	tell application "Finder"
		set window_list to every Finder window
		set paths to {}
		repeat with w in window_list
			set thePath to (POSIX path of (target of w as alias))
			copy thePath to end of paths
		end repeat
		return paths
	end tell`

	return utils.RunOsascript(script)
}

func (f Finder) Unpack(ws *WorkspaceConfig) error {
	if len(ws.Args) == 0 {
		return nil
	}
	return utils.OpenApp(ws.AppName, ws.Args...)
}

func (Finder) Quit(appName string) error {
	return utils.GracefulQuit(appName)
}
