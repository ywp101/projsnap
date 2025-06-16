package apps

import (
	"fmt"
	"workspace/utils"
)

type Browser struct {
}

func (f Browser) Pack(_, browserName string) ([]string, error) {
	browserScript := fmt.Sprintf(`
tell application "%s"
	set tabList to {}
	repeat with w in windows
		repeat with t in tabs of w
			set end of tabList to URL of t
		end repeat
	end repeat
	return tabList
end tell`, browserName)

	return utils.RunOsascript(browserScript)
}

func (Browser) Unpack(ws *WorkspaceConfig) error {
	return utils.OpenApp(ws.AppName, ws.Args...)
}

func (Browser) Quit(browserName string) error {
	quitScript := fmt.Sprintf(` tell application "%s" to close every window`, browserName)
	_, err := utils.RunOsascript(quitScript)
	return err
}
