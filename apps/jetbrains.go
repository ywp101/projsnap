package apps

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"workspace/utils"
)

type JetBrains struct {
}

func getJetBrainsIOOpenFiles(appName string) ([]string, error) {
	titles, err := utils.GetCurrenWindowsFile(appName)
	if err != nil {
		return nil, err
	}
	fileNames := make([]string, 0)
	for _, title := range titles {
		if title == "" {
			continue
		}
		tmp := strings.Split(title, " – ") // is e28093, "–" != "-"
		if len(tmp) != 2 {
			return nil, errors.New("parse JetBrains title error")
		}
		fileNames = append(fileNames, tmp[0])
	}
	return fileNames, nil
}

func (JetBrains) Pack(_, ideName string) ([]string, error) {
	projectNames, err := getJetBrainsIOOpenFiles(ideName)
	if err != nil {
		return nil, fmt.Errorf("getJetBrainsIOOpenFiles occur fail, err: %v\n", err)
	}
	xmlPath := "/Users/ryanye/Library/Application Support/JetBrains/GoLand2023.2/options/recentProjects.xml"
	data, err := os.ReadFile(xmlPath)
	if err != nil {
		fmt.Println("读取失败:", err)
		return nil, err
	}
	recentProjects := make(map[string]string)
	reg, err := regexp.Compile(`<entry key="(.*?)">`)
	if err != nil {
		return nil, err
	}
	results := reg.FindAllStringSubmatch(string(data), -1)
	for _, result := range results {
		recentProjects[filepath.Base(result[1])] = result[1]
	}
	openProjects := make([]string, 0)
	for _, pn := range projectNames {
		p := recentProjects[pn]
		expendedPath, _ := utils.ExpandUser(p)
		openProjects = append(openProjects, expendedPath)
	}

	return openProjects, nil
}

func (JetBrains) Unpack(ws *WorkspaceConfig) error {
	return utils.OpenApp(ws.AppName, ws.Args...)
}

func (JetBrains) Quit(appName string) error {
	return utils.GracefulQuit(appName)
}
