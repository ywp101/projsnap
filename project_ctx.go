package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"path/filepath"
	"projctx/apps"
	"projctx/utils"
	"strconv"
	"strings"
	"time"
)

type AppSnapshot struct {
	*apps.AppConfig
	*WindowInfo
}

type ProjectCtxSnapshot struct {
	SnapshotAlias string `json:"snapshot_alias"`
	SnapshotKey   string `json:"snapshot_key"`
	Ctime         int64  `json:"ctime"`
}

type ProjectCtxMeta struct {
	Snapshots map[string]ProjectCtxSnapshot `json:"snapshots"`
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
	db            *bolt.DB
	wm            *WindowManager
}

func NewWorkspace(opt *ProjectCtxOptions) *ProjectCtx {
	ws := &ProjectCtx{
		specSessions:  make(map[string]apps.AppPacker),
		generalPacker: apps.NormalPacker{},
		opt:           opt,
		meta:          &ProjectCtxMeta{Snapshots: make(map[string]ProjectCtxSnapshot)},
		wm:            NewWindowManager(),
	}
	LoadApplicationPlugins(ws)
	dbPath := filepath.Join(opt.configDir, "projctx.db")
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	ws.db = db
	return ws
}

func (w *ProjectCtx) Close() error {
	if w.db != nil {
		return w.db.Close()
	}
	return nil
}

func (w *ProjectCtx) Open() (err error) {
	if err = w.db.Update(func(tx *bolt.Tx) error {
		_, _ = tx.CreateBucketIfNotExists([]byte("ctx"))
		_, _ = tx.CreateBucketIfNotExists([]byte("snapshots"))
		return nil
	}); err != nil {
		return err
	}
	if err = w.loadMeta(); err != nil {
		return err
	}
	return w.wm.PreCheck()
}

func (w *ProjectCtx) RemoveSnapshots(aliasName string) error {
	snapshot, ok := w.meta.Snapshots[aliasName]
	if !ok {
		return fmt.Errorf("no found ctxID or aliasName: %s", aliasName)
	}
	// todo: 其他app产生的config也需要清理
	return w.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("ctx"))
		ssBucket := tx.Bucket([]byte("snapshots"))
		_ = ssBucket.Delete([]byte(aliasName))
		_ = b.Delete([]byte(snapshot.SnapshotKey))
		return nil
	})
}

func (w *ProjectCtx) ListSnapshots() map[string]ProjectCtxSnapshot {
	return w.meta.Snapshots
}

func (w *ProjectCtx) loadMeta() error {
	return w.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("snapshots"))
		c := b.Cursor()
		//ss := tx.Bucket([]byte("ctx"))
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			ps := ProjectCtxSnapshot{}
			if err := json.Unmarshal(v, &ps); err != nil {
				return err
			}
			w.meta.Snapshots[string(k)] = ps
			//fmt.Println(string(ss.Get([]byte(ps.SnapshotKey))))
		}
		return nil
	})
}

func (w *ProjectCtx) saveMeta(aliasName string, appSnapshots []AppSnapshot) (seq uint64, err error) {
	err = w.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("ctx"))
		seq, _ = b.NextSequence()
		ctxID := strconv.FormatUint(seq, 10)
		data, err := json.Marshal(appSnapshots)
		if err != nil {
			return err
		}
		if err := b.Put([]byte(ctxID), data); err != nil {
			return err
		}
		ps := ProjectCtxSnapshot{
			SnapshotAlias: aliasName,
			SnapshotKey:   ctxID,
			Ctime:         time.Now().Unix(),
		}
		if aliasName == "" {
			ps.SnapshotAlias = ctxID
		}
		ssData, err := json.Marshal(ps)
		if err != nil {
			return err
		}
		ssBucket := tx.Bucket([]byte("snapshots"))
		return ssBucket.Put([]byte(aliasName), ssData)
	})
	return
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

func (w *ProjectCtx) QuitAllApplication(appNames map[string]struct{}) {
	hasTerm := false
	for app := range appNames {
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
	if err := w.wm.TakeSnapshot(); err != nil {
		return false, err
	}

	appSnapshots := make([]AppSnapshot, 0)
	for app := range appNames {
		conf, err := w.GetPacker(app).Pack(w.opt.configDir, app)
		if err != nil {
			return false, fmt.Errorf("%s occur fail, err: %v", app, err)
		}
		// todo: save要关联正常，restore关联也要正常，现在是随机
		for i := range conf {
			wind, _ := w.wm.GetWindowInfo(app) // ignore error
			appSnapshots = append(appSnapshots, AppSnapshot{AppConfig: &conf[i], WindowInfo: wind})
		}
	}

	ctxID, err := w.saveMeta(aliasName, appSnapshots)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("SaveWorkSpace Success, ctxID: %d, alias: %s", ctxID, aliasName)

	if w.opt.quit {
		w.QuitAllApplication(appNames)
	}
	return true, nil
}

func (w *ProjectCtx) preloadWorkspace(aliasName string) []AppSnapshot {
	// alias
	snapshot, ok := w.meta.Snapshots[aliasName]
	if !ok {
		log.Fatal("no found ctxID or aliasName")
	}
	appSnapshots := make([]AppSnapshot, 0)

	_ = w.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("ctx"))
		data := b.Get([]byte(snapshot.SnapshotKey))
		return json.Unmarshal(data, &appSnapshots)
	})
	return appSnapshots
}

func (w *ProjectCtx) SwitchWorkspace(aliasName string) error {
	appSnapshots := w.preloadWorkspace(aliasName)
	nowAppNames, err := w.GetAllApplication()
	if err != nil {
		return err
	}
	for i, conf := range appSnapshots {
		log.Printf("[%d/%d] Opening %s, args: %v\n", i+1, len(appSnapshots), conf.AppName, conf.Args)
		_, running := nowAppNames[conf.AppName]
		if err := w.GetPacker(conf.AppName).Unpack(conf.AppConfig, running); err != nil {
			return err
		}
		delete(nowAppNames, conf.AppName)
	}
	for app := range nowAppNames {
		log.Printf("close no use app: %s\n", app)
		_ = w.GetPacker(app).Quit(app)
	}
	time.Sleep(3 * time.Second)
	if err := w.wm.TakeSnapshot(); err != nil {
		return err
	}
	// restore windows
	for _, conf := range appSnapshots {
		_ = w.wm.RestoreWindow(conf.WindowInfo)
	}
	return nil
}

func (w *ProjectCtx) LoadWorkspace(aliasName string) error {
	appSnapshots := w.preloadWorkspace(aliasName)
	for i, conf := range appSnapshots {
		log.Printf("[%d/%d] Opening %s, args: %v\n", i+1, len(appSnapshots), conf.AppName, conf.Args)
		if err := w.GetPacker(conf.AppName).Unpack(conf.AppConfig, false); err != nil {
			return err
		}
	}
	time.Sleep(3 * time.Second)
	// get current opened windows
	if err := w.wm.TakeSnapshot(); err != nil {
		return err
	}
	// restore windows
	for _, conf := range appSnapshots {
		_ = w.wm.RestoreWindow(conf.WindowInfo)
	}

	return nil
}

func (w *ProjectCtx) GetAllApplication() (map[string]struct{}, error) {
	allApp, err := utils.RunOsascript(`
	tell application "System Events"
		get name of (processes where background only is false)
	end tell`)
	if err != nil {
		return nil, err
	}
	result := make(map[string]struct{})
	for _, app := range allApp {
		result[app] = struct{}{}
	}
	return result, nil
}
