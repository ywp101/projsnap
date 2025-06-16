package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"workspace/apps"
	"workspace/utils"
)

type WorkspaceMeta struct {
	Snapshots map[string]string `json:"snapshots"` // aliasName -> workspace.json
}

type WorkspaceOptions struct {
	quit      bool
	configDir string
}

type Workspace struct {
	specSessions  map[string]apps.AppSessionPacker
	generalPacker apps.AppSessionPacker
	opt           *WorkspaceOptions
	meta          *WorkspaceMeta
}

func NewWorkspace(opt *WorkspaceOptions) *Workspace {
	ws := &Workspace{
		specSessions:  make(map[string]apps.AppSessionPacker),
		generalPacker: apps.NormalPacker{},
		opt:           opt,
		meta:          &WorkspaceMeta{Snapshots: make(map[string]string)},
	}
	LoadApplicationPlugins(ws)
	if err := ws.loadMeta(); err != nil {
		log.Fatal(err)
	}
	return ws
}

func (w *Workspace) getMetaFilePath() string {
	return filepath.Join(w.opt.configDir, "meta.json")
}

func (w *Workspace) ListSnapshots() map[string]string {
	return w.meta.Snapshots
}

func (w *Workspace) loadMeta() error {
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.Mkdir(configDir, 0755); err != nil {
			return err
		}
	}

	metaPath := w.getMetaFilePath()
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		w.meta = &WorkspaceMeta{Snapshots: make(map[string]string)}
		return nil
	}
	fd, err := os.Open(metaPath)
	if err != nil {
		return err
	}
	defer fd.Close()
	return json.NewDecoder(fd).Decode(w.meta)
}

func (w *Workspace) saveMeta(aliasName, wsFilePath string) error {
	w.meta.Snapshots[aliasName] = wsFilePath

	metaPath := w.getMetaFilePath()
	fd, err := os.Create(metaPath)
	if err != nil {
		return err
	}
	defer fd.Close()
	return json.NewEncoder(fd).Encode(w.meta)
}

func (w *Workspace) RegisterApplication(name string, app apps.AppSessionPacker) {
	w.specSessions[strings.ToLower(name)] = app
}

func (w *Workspace) GetPacker(appName string) apps.AppSessionPacker {
	appName = strings.ToLower(appName)
	if packer, ok := w.specSessions[appName]; ok {
		return packer
	}
	return w.generalPacker
}

func (w *Workspace) QuitAllApplication(appNames []string) {
	hasTerm := false
	for _, app := range appNames {
		if app != "iTerm2" {
			log.Printf("quit %s, err: %v", app, w.GetPacker(app).Quit(app))
		} else {
			hasTerm = true
		}
	}
	// todo: hard code
	if hasTerm {
		_ = w.GetPacker("iTerm2").Quit("iTerm2")
	}
}

func (w *Workspace) SaveWorkspace(aliasName string) (bool, error) {
	appNames, err := w.GetAllApplication()
	if err != nil {
		return false, err
	}

	appNames = RemoveInWhiteList(appNames)
	wsConfigs := make([]apps.WorkspaceConfig, 0)
	hashInput := ""
	for _, app := range appNames {
		hashInput += app
		conf, err := w.GetPacker(app).Pack(w.opt.configDir, app)
		if err != nil {
			return false, fmt.Errorf("%s occur fail, err: %v", app, err)
		}
		wsConfigs = append(wsConfigs, apps.WorkspaceConfig{AppName: app, Args: conf})
	}

	fname := utils.Hash(hashInput + time.Now().String())
	fd, err := os.Create(filepath.Join(w.opt.configDir, fname+".json"))
	if err != nil {
		log.Fatalf("Create workspace file fail, err:%v\n", err)
	}
	defer fd.Close()
	if err := json.NewEncoder(fd).Encode(wsConfigs); err != nil {
		log.Fatal(err)
	}

	if aliasName == "" {
		aliasName = fname
	}
	if err := w.saveMeta(aliasName, fname); err != nil {
		log.Fatal(err)
	}
	log.Printf("SaveWorkSpace Success, ctxID: %s, alias: %s", fname, aliasName)

	if w.opt.quit {
		w.QuitAllApplication(appNames)
	}
	return true, nil
}

func (w *Workspace) LoadWorkspace(aliasName string) error {
	// alias
	ctxVersion, ok := w.meta.Snapshots[aliasName]
	if !ok {
		return errors.New("no found ctxID or aliasName")
	}
	configPath := filepath.Join(w.opt.configDir, ctxVersion+".json")
	fd, err := os.Open(configPath)
	if err != nil {
		log.Fatal(err)
	}
	wsConfigs := make([]apps.WorkspaceConfig, 0)
	if err = json.NewDecoder(fd).Decode(&wsConfigs); err != nil {
		log.Fatal(err)
	}
	for i, conf := range wsConfigs {
		log.Printf("[%d/%d] Opening %s, args: %v\n", i+1, len(wsConfigs), conf.AppName, conf.Args)
		if err := w.GetPacker(conf.AppName).Unpack(&conf); err != nil {
			return err
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
