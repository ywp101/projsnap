package apps

import (
	"path/filepath"
	"projctx/utils"
	"strings"
)

var (
	drawIOConfigPath = "~/Library/Application Support/draw.io/Local Storage/leveldb"
	drawIOAppName    = "draw.io"
)

type DrawIO struct {
}

func getDrawIOOpenFiles() ([]string, error) {
	titles, err := utils.GetCurrenWindowsFile(drawIOAppName)
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

func (d DrawIO) Pack(_, _ string) ([]string, error) {
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

func (d DrawIO) Unpack(ws *AppConfig, running bool) error {
	if running {
		_ = d.Quit(ws.AppName)
	}
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

func (d DrawIO) Quit(appName string) error {
	return utils.GracefulQuit(appName)
}
