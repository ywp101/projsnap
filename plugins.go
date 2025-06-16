package main

import (
	"strings"
	"workspace/apps"
)

func LoadApplicationPlugins(ws *Workspace) {
	ws.RegisterApplication("Finder", apps.Finder{})
	ws.RegisterApplication("Microsoft Edge", apps.Browser{})
	ws.RegisterApplication("draw.io", apps.DrawIO{})
	ws.RegisterApplication("Obsidian", apps.Obsidian{})
	ws.RegisterApplication("iterm2", apps.Iterm2{})
	ws.RegisterApplication("goland", apps.JetBrains{})
}

func RemoveInWhiteList(appNames []string) []string {
	whiteList := make(map[string]bool)
	whiteList["goland"] = true
	whiteList["iterm2"] = true

	result := make([]string, 0)
	for _, app := range appNames {
		if _, ok := whiteList[strings.ToLower(app)]; !ok {
			result = append(result, app)
		}
	}
	return result
}
