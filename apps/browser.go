package apps

import (
	"fmt"
	"projctx/utils"
	"strings"
)

type Browser struct {
}

func (b Browser) Pack(_, browserName string) ([]AppConfig, error) {
	browserScript := fmt.Sprintf(`
tell application "%s"
	set winList to {}
	repeat with w in windows
		set tabList to {}
		repeat with t in tabs of w
			set end of tabList to URL of t
		end repeat
		set end of winList to tabList & "\n"
	end repeat
	return winList
end tell`, browserName)

	tabs := make([]AppConfig, 0)
	err := utils.RunOsascriptWithSplit(browserScript, func(output string) error {
		windows := strings.Split(output, "\n")
		for _, wind := range windows {
			if wind == "" {
				continue
			}
			tmp := AppConfig{AppName: browserName, Args: make([]string, 0)}
			for _, tab := range strings.Split(wind, ",") {
				v := strings.TrimSpace(tab)
				if v == "" {
					continue
				}
				tmp.Args = append(tmp.Args, v)
			}
			tabs = append(tabs, tmp)
		}
		return nil
	})
	return tabs, err
}

func (b Browser) Unpack(ws *AppConfig, running bool) error {
	if running {
		_ = b.Quit(ws.AppName)
	}
	openArgs := make([]string, 0)
	openArgs = append(openArgs, "-n", "--args")
	for _, tab := range ws.Args {
		openArgs = append(openArgs, "--new-window", tab)
	}
	err := utils.OpenApp(ws.AppName, openArgs...)
	if err != nil {
		return err
	}
	return nil
}

func (b Browser) Quit(browserName string) error {
	quitScript := fmt.Sprintf(` tell application "%s" to close every window`, browserName)
	_, err := utils.RunOsascript(quitScript)
	return err
}
