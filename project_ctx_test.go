package main

import (
	"log"
	"testing"
)

func TestNewWorkspacePack(t *testing.T) {
	ws := NewWorkspace(&ProjectCtxOptions{
		configDir: configDir,
	})
	if err := ws.Open(); err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	_, _ = ws.GetPacker("Microsoft Edge").Pack("", "Microsoft Edge")
}

func TestNewWorkspaceUnPack(t *testing.T) {
	ws := NewWorkspace(&ProjectCtxOptions{
		configDir: configDir,
	})
	if err := ws.Open(); err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	appName := "Microsoft Edge"
	appSnapshots := ws.preloadWorkspace("fuck")
	for _, sshot := range appSnapshots {
		if sshot.AppName == appName {
			_ = ws.GetPacker(appName).Unpack(sshot.AppConfig, true)
		}
	}
}
