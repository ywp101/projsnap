package apps

import (
	"fmt"
	"os"
	"path/filepath"
	"projsnap/utils"
	"regexp"
)

type JetBrains struct {
}

func getJetBrainsIOOpenFiles(appName string) ([]string, error) {
	titles, err := utils.GetCurrenWindowsFile(appName)
	if err != nil {
		return nil, err
	}
	return utils.SliceSplit(titles, " – ", 2, 0) // is e28093, "–" != "-"
}

func ReadRecentProjectFile(ideName string) ([]byte, error) {
	pattern, _ := utils.ExpandUser("~/Library/Application Support/JetBrains/*/options/recentProjects.xml")
	tmp, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	matched := utils.FindMatched(tmp, ideName)
	if len(matched) != 1 {
		return nil, fmt.Errorf("find matched file occur fail, matched: %v\n", matched)
	}
	return os.ReadFile(matched[0])
}

func (j JetBrains) Pack(_, ideName string) ([]AppConfig, error) {
	projectNames, err := getJetBrainsIOOpenFiles(ideName)
	if err != nil {
		return nil, fmt.Errorf("getJetBrainsIOOpenFiles occur fail, err: %v\n", err)
	}
	data, err := ReadRecentProjectFile(ideName)
	if err != nil {
		fmt.Println("读取失败:", err)
		return nil, err
	}
	// read recent all projects from xml file
	recentProjects := make(map[string]string)
	reg, err := regexp.Compile(`<entry key="(.*?)">`)
	if err != nil {
		return nil, err
	}
	results := reg.FindAllStringSubmatch(string(data), -1)
	for _, result := range results {
		recentProjects[filepath.Base(result[1])] = result[1]
	}
	// found opened project
	openProjects := make([]string, 0)
	for _, pn := range projectNames {
		p := recentProjects[pn]
		expendedPath, _ := utils.ExpandUser(p)
		openProjects = append(openProjects, expendedPath)
	}

	return NewAppConfigsWithArgs(ideName, openProjects), nil
}

func (j JetBrains) Unpack(ws *AppConfig, running bool) error {
	if running {
		_ = j.Quit(ws.AppName)
	}
	return utils.OpenMultiApp(ws.AppName, ws.Args...)
}

func (j JetBrains) Quit(appName string) error {
	return utils.GracefulQuit(appName)
}
