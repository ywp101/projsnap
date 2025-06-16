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

type ProjectCtxMeta struct {
	Snapshots map[string]string `json:"snapshots"` // aliasName -> workspace.json
}

type ProjectCtxOptions struct {
	quit      bool
	configDir string
}

type ProjectCtx struct {
	specSessions  map[string]apps.AppPacker
	generalPacker apps.AppPacker
	opt           *ProjectCtxOptions
	meta          *ProjectCtxMeta
}

func NewWorkspace(opt *ProjectCtxOptions) *ProjectCtx {
	ws := &ProjectCtx{
		specSessions:  make(map[string]apps.AppPacker),
		generalPacker: apps.NormalPacker{},
		opt:           opt,
		meta:          &ProjectCtxMeta{Snapshots: make(map[string]string)},
	}
	LoadApplicationPlugins(ws)
	if err := ws.loadMeta(); err != nil {
		log.Fatal(err)
	}
	return ws
}

func (w *ProjectCtx) getMetaFilePath() string {
	return filepath.Join(w.opt.configDir, "meta.json")
}

func (w *ProjectCtx) RemoveSnapshots(aliasName string) error {
	ctxVersion, ok := w.meta.Snapshots[aliasName]
	if !ok {
		return fmt.Errorf("no found ctxID or aliasName: %s", aliasName)
	}
	// todo: 其他app产生的config也需要清理
	delete(w.meta.Snapshots, aliasName)
	if err := os.Remove(filepath.Join(w.opt.configDir, ctxVersion+".json")); err != nil {
		return err
	}
	return w.saveMeta("", "")
}

func (w *ProjectCtx) ListSnapshots() map[string]string {
	return w.meta.Snapshots
}

func (w *ProjectCtx) loadMeta() error {
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.Mkdir(configDir, 0755); err != nil {
			return err
		}
	}

	metaPath := w.getMetaFilePath()
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		w.meta = &ProjectCtxMeta{Snapshots: make(map[string]string)}
		return nil
	}
	fd, err := os.Open(metaPath)
	if err != nil {
		return err
	}
	defer fd.Close()
	return json.NewDecoder(fd).Decode(w.meta)
}

func (w *ProjectCtx) saveMeta(aliasName, wsFilePath string) error {
	if aliasName != "" && wsFilePath != "" {
		w.meta.Snapshots[aliasName] = wsFilePath
	}

	metaPath := w.getMetaFilePath()
	fd, err := os.Create(metaPath)
	if err != nil {
		return err
	}
	defer fd.Close()
	return json.NewEncoder(fd).Encode(w.meta)
}

func (w *ProjectCtx) RegisterApplication(name string, app apps.AppPacker) {
	w.specSessions[strings.ToLower(name)] = app
}

func (w *ProjectCtx) GetPacker(appName string) apps.AppPacker {
	appName = strings.ToLower(appName)
	if packer, ok := w.specSessions[appName]; ok {
		return packer
	}
	return w.generalPacker
}

func (w *ProjectCtx) QuitAllApplication(appNames []string) {
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

func (w *ProjectCtx) SaveWorkspace(aliasName string) (bool, error) {
	appNames, err := w.GetAllApplication()
	if err != nil {
		return false, err
	}

	appNames = RemoveInWhiteList(appNames)
	wsConfigs := make([]apps.AppConfig, 0)
	hashInput := ""
	for _, app := range appNames {
		hashInput += app
		conf, err := w.GetPacker(app).Pack(w.opt.configDir, app)
		if err != nil {
			return false, fmt.Errorf("%s occur fail, err: %v", app, err)
		}
		wsConfigs = append(wsConfigs, apps.AppConfig{AppName: app, Args: conf})
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

func (w *ProjectCtx) LoadWorkspace(aliasName string) error {
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
	wsConfigs := make([]apps.AppConfig, 0)
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

func (w *ProjectCtx) GetAllApplication() ([]string, error) {
	return utils.RunOsascript(`
	tell application "System Events"
		get name of (processes where background only is false)
	end tell`)
}
