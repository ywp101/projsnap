package apps

import (
	"fmt"
	"os"
	"path/filepath"
	"projctx/utils"
	"time"
)

type Iterm2 struct {
}

func (Iterm2) Pack(_, appName string) ([]AppConfig, error) {
	iterm2File := filepath.Join(os.TempDir(), ".iterm2.txt")
	_ = os.Remove(iterm2File)
	script := fmt.Sprintf(`tell application "iTerm"
  set output to ""
  repeat with w in windows
    repeat with t in tabs of w
      repeat with s in sessions of t
        tell s to write text "pwd >> %s"
      end repeat
    end repeat
  end repeat
  return output
end tell`, iterm2File)
	_, err := utils.RunOsascript(script)
	if err != nil {
		return nil, err
	}
	time.Sleep(1 * time.Second) // wait for iterm2 to write
	defer os.Remove(iterm2File)
	result, err := utils.ReadFileToStringList(iterm2File)
	return NewAppConfigsWithArgs(appName, result), err
}

func (Iterm2) Unpack(ws *AppConfig, _ bool) error {
	return utils.OpenApp("iterm", ws.Args...)
}

func (Iterm2) Quit(appName string) error {
	return utils.GracefulQuit(appName)
}
