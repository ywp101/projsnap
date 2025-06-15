package main

import (
	"fmt"
	"workspace/apps"
)

func main() {

	ws := NewWorkspace()
	//ws.RegisterApplication("Finder", apps.Finder{})
	//ws.RegisterApplication("Microsoft Edge", apps.Browser{})
	//ws.RegisterApplication("draw.io", apps.DrawIO{})
	//ws.RegisterApplication("Obsidian", apps.Obsidian{})
	ws.RegisterApplication("iterm2", apps.Iterm2{})

	//ws.SaveWorkspace()
	fmt.Println(ws.LoadWorkspace("./workspace.json"))
}
