package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"workspace/apps"
	"workspace/utils"
)

type Workspace struct {
	specSessions  map[string]apps.AppSessionPacker
	generalPacker apps.AppSessionPacker
}

func NewWorkspace() *Workspace {
	return &Workspace{
		specSessions:  make(map[string]apps.AppSessionPacker),
		generalPacker: apps.NormalPacker{},
	}
}

func (w *Workspace) RegisterApplication(name string, app apps.AppSessionPacker) {
	w.specSessions[strings.ToLower(name)] = app
}

func (w *Workspace) SaveWorkspace() (bool, error) {
	appNames, err := w.GetAllApplication()
	if err != nil {
		return false, err
	}
	wsConfigs := make([]apps.WorkspaceConfig, 0)
	for _, app := range appNames {
		session := w.generalPacker
		if spec, ok := w.specSessions[strings.ToLower(app)]; ok {
			session = spec
		}

		conf, err := session.Pack(app)
		if err != nil {
			return false, err
		}
		wsConfigs = append(wsConfigs, apps.WorkspaceConfig{AppName: app, Args: conf})
	}
	buf, err := json.Marshal(wsConfigs)
	if err != nil {
		return false, err
	}
	fd, err := os.Create("./workspace.json")
	fd.Write(buf)
	fd.Sync()
	defer fd.Close()
	return true, nil
}

func (w *Workspace) LoadWorkspace(configPath string) error {
	wsConfigs := make([]apps.WorkspaceConfig, 0)
	fd, err := os.Open(configPath)
	if err != nil {
		log.Fatal(err)
	}
	err = json.NewDecoder(fd).Decode(&wsConfigs)
	if err != nil {
		log.Fatal(err)
	}
	// todo: 指定app path
	for _, conf := range wsConfigs {
		if session, ok := w.specSessions[strings.ToLower(conf.AppName)]; ok {
			if err := session.Unpack(&conf); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *Workspace) GetAllApplication() ([]string, error) {
	return utils.RunOsascript(`
	tell application "System Events"
		get name of (processes where background only is false)
	end tell`)
}
