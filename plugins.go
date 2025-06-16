package main

import "workspace/apps"

func LoadApplicationPlugins(ws *Workspace) {
	ws.RegisterApplication("Finder", apps.Finder{})
	ws.RegisterApplication("Microsoft Edge", apps.Browser{})
	ws.RegisterApplication("draw.io", apps.DrawIO{})
	ws.RegisterApplication("Obsidian", apps.Obsidian{})
	ws.RegisterApplication("iterm2", apps.Iterm2{})
	ws.RegisterApplication("goland", apps.JetBrains{})
}
