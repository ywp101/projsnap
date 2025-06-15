package apps

import (
	"path/filepath"
	"strings"
	"workspace/utils"
)

var (
	drawIOConfigPath = "~/Library/Application Support/draw.io/Local Storage/leveldb"
)

type DrawIO struct {
}

func getCurrenDrawIOFile() ([]string, error) {
	script := `
	tell application "System Events"
		set appName to "draw.io"
		set winTitles to {}
		repeat with w in windows of application process appName
			set end of winTitles to name of w
		end repeat
		return winTitles
	end tell`
	return utils.RunOsascript(script)
}

func getDrawIOOpenFiles() ([]string, error) {
	titles, err := getCurrenDrawIOFile()
	if err != nil {
		return nil, err
	}
	fileNames := make([]string, 0)
	suffix := ".drawio"
	for _, title := range titles {
		i := strings.Index(title, suffix)
		if i != -1 {
			fileNames = append(fileNames, title[:i+len(suffix)])
		}
	}
	return fileNames, nil
}

func (DrawIO) Pack(_ string) ([]string, error) {
	fileNames, err := getDrawIOOpenFiles()
	if err != nil {
		return nil, err
	}
	if err := utils.GracefulQuit("draw.io"); err != nil {
		return nil, err
	}

	recentFiles, err := utils.ReadRecentFiles(drawIOConfigPath)
	if err != nil {
		return nil, err
	}
	filePaths := make([]string, 0)
	for _, fileName := range fileNames {
		for _, filePath := range recentFiles {
			if strings.HasSuffix(filePath, fileName) {
				filePaths = append(filePaths, filePath)
			}
		}
	}
	return filePaths, nil
}

func (DrawIO) Unpack(ws *WorkspaceConfig) error {
	taskMap := make(map[string]bool)
	return utils.OpenMultiAppByRetry(func() ([]string, error) {
		doneArgs, err := getDrawIOOpenFiles()
		if err != nil {
			return nil, err
		}
		for _, arg := range doneArgs {
			taskMap[arg] = true
		}

		failedArgs := make([]string, 0)
		for _, arg := range ws.Args {
			if _, ok := taskMap[filepath.Base(arg)]; !ok {
				failedArgs = append(failedArgs, arg)
			}
		}
		return failedArgs, nil
	}, ws.AppName, ws.Args...)
}
