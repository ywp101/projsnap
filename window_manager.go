package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"projctx/utils"
	"strconv"
)

type Rect struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}

type WindowInfo struct {
	App       string `json:"app"`
	Title     string `json:"title"`
	Frame     Rect   `json:"frame"`
	SpaceID   int    `json:"space"`
	DisplayID int    `json:"display"`
	WindowID  int    `json:"id"`
	Pid       int    `json:"pid"`
}

type WindowManager struct {
	savedWindows []WindowInfo
	readedWindow []int
}

func NewWindowManager() *WindowManager {
	return &WindowManager{
		savedWindows: make([]WindowInfo, 0),
	}
}

func (wm *WindowManager) PreCheck() error {
	_, err := exec.LookPath("yabai")
	if err != nil {
		return errors.New("yabai not found in PATH, need to `brew install koekeishiya/formulae/yabai`")
	}
	if !IsYabaiRunning() {
		return errors.New("yabai is not running, you should run `yabai --start-service`")
	}
	return nil
}

func IsYabaiRunning() bool {
	cmd := exec.Command("pgrep", "yabai")
	err := cmd.Run()
	return err == nil
}

func (wm *WindowManager) TakeSnapshot() error {
	cmd := exec.Command("yabai", "-m", "query", "--windows")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("yabai query failed: %w", err)
	}

	var windows []WindowInfo
	if err := json.Unmarshal(output, &windows); err != nil {
		return fmt.Errorf("unmarshal failed: %w", err)
	}

	wm.savedWindows = windows
	wm.readedWindow = make([]int, len(windows))
	return nil
}

func (wm *WindowManager) GetWindowInfo(appName string) (*WindowInfo, error) {
	win, err := wm.GetWindowFromName(appName)
	if err == nil {
		return win, nil
	}
	// no english app name
	pids, err := utils.GetPIDFromAppName(appName)
	if err != nil {
		return nil, err
	}
	for _, pid := range pids {
		for _, win := range wm.savedWindows {
			if win.Pid == pid {
				return &win, nil
			}
		}
	}
	return nil, fmt.Errorf("window not found for app: %s", appName)
}

func (wm *WindowManager) GetWindowFromName(appName string) (*WindowInfo, error) {
	for i := len(wm.savedWindows) - 1; i >= 0; i-- {
		win := wm.savedWindows[i]
		if win.App == appName && wm.readedWindow[i] == 0 {
			wm.readedWindow[i] = 1
			return &win, nil
		}
	}
	return nil, fmt.Errorf("window not found for app name: %s", appName)
}

func (wm *WindowManager) RestoreWindow(win *WindowInfo) error {
	// ignore
	if win == nil {
		return nil
	}
	curWin, err := wm.GetWindowFromName(win.App)
	if err != nil {
		return err
	}
	curWinID := strconv.Itoa(curWin.WindowID)
	log.Printf("curWinID:%s, win: %v\n", curWinID, win)

	// 尝试移动到 space
	if out, err := exec.Command("yabai", "-m", "window", curWinID, "--space", strconv.Itoa(win.SpaceID)).CombinedOutput(); err != nil {
		log.Printf("move to space %d fail: %v, out:%v\n", win.SpaceID, err, string(out))
	}

	// 恢复位置大小
	frame := win.Frame
	pos := fmt.Sprintf("abs:%d:%d", int(frame.X), int(frame.Y))
	if out, err := exec.Command("yabai", "-m", "window", curWinID, "--move", pos).CombinedOutput(); err != nil {
		log.Printf("move pos to %s failed for window %d: %v, output: %s\n", pos, win.WindowID, err, string(out))
	}

	windSize := fmt.Sprintf("--resize abs:%d:%d", int(frame.W), int(frame.H))
	if out, err := exec.Command("yabai", "-m", "window", curWinID, "--resize", windSize).CombinedOutput(); err != nil {
		log.Printf("resize to %s failed for window %d: %v, output: %s\n", pos, win.WindowID, err, string(out))
	}

	return nil
}
