package apps

import (
	"projsnap/utils"
)

type Finder struct {
}

func (f Finder) Pack(_, appName string) ([]AppConfig, error) {
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

	tmp, err := utils.RunOsascript(script)
	return NewAppConfigsWithArgs(appName, tmp), err
}

func (f Finder) Unpack(ws *AppConfig, running bool) error {
	if len(ws.Args) == 0 {
		return nil
	}
	if running {
		_ = f.Quit(ws.AppName)
	}
	return utils.OpenApp(ws.AppName, ws.Args...)
}

func (f Finder) Quit(appName string) error {
	return utils.GracefulQuit(appName)
}
