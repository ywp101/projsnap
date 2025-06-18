package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"path/filepath"
	"projsnap/apps"
	"projsnap/utils"
	"strconv"
	"strings"
	"time"
)

var (
	manifestBucketName  = []byte("manifest")
	SnapshotsBucketName = []byte("snapshots")
)

type AppSnapshot struct {
	*apps.AppConfig
	*WindowInfo
}

type ProjSnapManifest struct {
	SnapshotName string `json:"snapshot_name"`
	SnapshotKey  string `json:"snapshot_key"`
	Ctime        int64  `json:"ctime"`
}

type ProjSnapMeta struct {
	ManifestSnapshots map[string]ProjSnapManifest `json:"manifest_snapshots"`
}

type ProjSnapOptions struct {
	quit      bool
	configDir string
}

type ProjSnapMaster struct {
	specPackers   map[string]apps.AppPacker
	generalPacker apps.AppPacker
	opt           *ProjSnapOptions
	meta          *ProjSnapMeta
	db            *bolt.DB
	wm            *WindowManager
}

func NewWorkspace(opt *ProjSnapOptions) *ProjSnapMaster {
	ws := &ProjSnapMaster{
		specPackers:   make(map[string]apps.AppPacker),
		generalPacker: apps.NormalPacker{},
		opt:           opt,
		meta:          &ProjSnapMeta{ManifestSnapshots: make(map[string]ProjSnapManifest)},
		wm:            NewWindowManager(),
	}
	LoadApplicationPlugins(ws)
	dbPath := filepath.Join(opt.configDir, "projsnap.db")
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	ws.db = db
	return ws
}

func (w *ProjSnapMaster) Close() error {
	if w.db != nil {
		return w.db.Close()
	}
	return nil
}

func (w *ProjSnapMaster) Open() (err error) {
	if err = w.db.Update(func(tx *bolt.Tx) error {
		_, _ = tx.CreateBucketIfNotExists(manifestBucketName)
		_, _ = tx.CreateBucketIfNotExists(SnapshotsBucketName)
		return nil
	}); err != nil {
		return err
	}
	if err = w.loadManifest(); err != nil {
		return err
	}
	return w.wm.PreCheck()
}

func (w *ProjSnapMaster) RemoveSnapshots(snapName string) error {
	snapshot, ok := w.meta.ManifestSnapshots[snapName]
	if !ok {
		return fmt.Errorf("no found snapName: %s", snapName)
	}
	return w.db.Update(func(tx *bolt.Tx) error {
		manifest := tx.Bucket(manifestBucketName)
		_ = manifest.Delete([]byte(snapName))

		snap := tx.Bucket(SnapshotsBucketName)
		_ = snap.Delete([]byte(snapshot.SnapshotKey))
		return nil
	})
}

func (w *ProjSnapMaster) ListSnapshots() map[string]ProjSnapManifest {
	return w.meta.ManifestSnapshots
}

func (w *ProjSnapMaster) loadManifest() error {
	return w.db.View(func(tx *bolt.Tx) error {
		manifest := tx.Bucket(manifestBucketName)
		c := manifest.Cursor()
		//ss := tx.Bucket([]byte("ctx"))
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			ps := ProjSnapManifest{}
			if err := json.Unmarshal(v, &ps); err != nil {
				return err
			}
			w.meta.ManifestSnapshots[string(k)] = ps
			//fmt.Println(string(ss.Get([]byte(ps.SnapshotKey))))
		}
		return nil
	})
}

func (w *ProjSnapMaster) dumpProjSnapshot(snapName string, appSnapshots []AppSnapshot) (seq uint64, err error) {
	oldSnap, _ := w.meta.ManifestSnapshots[snapName]
	err = w.db.Update(func(tx *bolt.Tx) error {
		snap := tx.Bucket(SnapshotsBucketName)
		seq, _ = snap.NextSequence()
		curSnapID := strconv.FormatUint(seq, 10)

		// delete old snapshot
		_ = snap.Delete([]byte(oldSnap.SnapshotKey))

		// save new snapshot
		data, err := json.Marshal(appSnapshots)
		if err != nil {
			return err
		}
		if err := snap.Put([]byte(curSnapID), data); err != nil {
			return err
		}

		// save manifest
		manifest := tx.Bucket([]byte("snapshots"))
		ps := ProjSnapManifest{
			SnapshotName: snapName,
			SnapshotKey:  curSnapID,
			Ctime:        time.Now().Unix(),
		}
		ssData, err := json.Marshal(ps)
		if err != nil {
			return err
		}
		return manifest.Put([]byte(snapName), ssData)
	})
	return
}

func (w *ProjSnapMaster) RegisterApplication(name string, app apps.AppPacker) {
	w.specPackers[strings.ToLower(name)] = app
}

func (w *ProjSnapMaster) GetPacker(appName string) apps.AppPacker {
	appName = strings.ToLower(appName)
	if packer, ok := w.specPackers[appName]; ok {
		return packer
	}
	return w.generalPacker
}

func (w *ProjSnapMaster) quitAllApplication(appNames map[string]struct{}) {
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

func (w *ProjSnapMaster) SaveSnapshot(snapName string) (bool, error) {
	appNames, err := w.getAllApplication()
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

	ctxID, err := w.dumpProjSnapshot(snapName, appSnapshots)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("SaveWorkSpace Success, ctxID: %d, alias: %s", ctxID, snapName)

	if w.opt.quit {
		w.quitAllApplication(appNames)
	}
	return true, nil
}

func (w *ProjSnapMaster) loadSnapshot(snapName string) []AppSnapshot {
	// alias
	snapshot, ok := w.meta.ManifestSnapshots[snapName]
	if !ok {
		log.Fatal("no found snapName")
	}
	appSnapshots := make([]AppSnapshot, 0)

	_ = w.db.View(func(tx *bolt.Tx) error {
		snap := tx.Bucket(SnapshotsBucketName)
		data := snap.Get([]byte(snapshot.SnapshotKey))
		return json.Unmarshal(data, &appSnapshots)
	})
	return appSnapshots
}

func (w *ProjSnapMaster) openAppFromSnapshot(appSnapshots []AppSnapshot, realRunning map[string]struct{}) error {
	for i, conf := range appSnapshots {
		log.Printf("[%d/%d] Opening %s, args: %v\n", i+1, len(appSnapshots), conf.AppName, conf.Args)
		_, running := realRunning[conf.AppName]
		if err := w.GetPacker(conf.AppName).Unpack(conf.AppConfig, running); err != nil {
			return err
		}
	}
	return nil
}

func (w *ProjSnapMaster) SwitchSnapshot(snapName string) error {
	appSnapshots := w.loadSnapshot(snapName)
	realRunning, err := w.getAllApplication()
	if err != nil {
		return err
	}
	// open all
	if err := w.openAppFromSnapshot(appSnapshots, realRunning); err != nil {
		return err
	}
	// close other app
	for app := range realRunning {
		shouldClose := true
		for _, conf := range appSnapshots {
			if conf.AppName == app {
				shouldClose = false
				break
			}
		}
		if shouldClose {
			log.Printf("Closing %s\n", app)
			_ = w.GetPacker(app).Quit(app)
		}
	}
	// wait
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

func (w *ProjSnapMaster) RestoreSnapshot(snapName string) error {
	appSnapshots := w.loadSnapshot(snapName)
	// open app, ignore current whether is opened
	if err := w.openAppFromSnapshot(appSnapshots, map[string]struct{}{}); err != nil {
		return err
	}
	// wait app
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

func (w *ProjSnapMaster) getAllApplication() (map[string]struct{}, error) {
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
