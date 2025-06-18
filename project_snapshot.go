package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
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
	psm := &ProjSnapMaster{
		specPackers:   make(map[string]apps.AppPacker),
		generalPacker: apps.NormalPacker{},
		opt:           opt,
		meta:          &ProjSnapMeta{ManifestSnapshots: make(map[string]ProjSnapManifest)},
		wm:            NewWindowManager(),
	}
	LoadApplicationPlugins(psm)
	return psm
}

func (psm *ProjSnapMaster) Close() error {
	if psm.db != nil {
		return psm.db.Close()
	}
	return nil
}

func (psm *ProjSnapMaster) Open() (err error) {
	if _, err := os.Stat(psm.opt.configDir); os.IsNotExist(err) {
		if err = os.MkdirAll(psm.opt.configDir, 0755); err != nil {
			return err
		}
	}

	// open db
	dbPath := filepath.Join(psm.opt.configDir, "projsnap.db")
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	psm.db = db

	// create bucket
	if err = psm.db.Update(func(tx *bolt.Tx) error {
		_, _ = tx.CreateBucketIfNotExists(manifestBucketName)
		_, _ = tx.CreateBucketIfNotExists(SnapshotsBucketName)
		return nil
	}); err != nil {
		return err
	}
	if err = psm.loadManifest(); err != nil {
		return err
	}
	// check yabai
	return psm.wm.PreCheck()
}

func (psm *ProjSnapMaster) RemoveSnapshots(snapName string) error {
	snapshot, ok := psm.meta.ManifestSnapshots[snapName]
	if !ok {
		return fmt.Errorf("no found snapName: %s", snapName)
	}
	return psm.db.Update(func(tx *bolt.Tx) error {
		manifest := tx.Bucket(manifestBucketName)
		_ = manifest.Delete([]byte(snapName))

		snap := tx.Bucket(SnapshotsBucketName)
		_ = snap.Delete([]byte(snapshot.SnapshotKey))
		return nil
	})
}

func (psm *ProjSnapMaster) ListSnapshots() map[string]ProjSnapManifest {
	return psm.meta.ManifestSnapshots
}

func (psm *ProjSnapMaster) loadManifest() error {
	return psm.db.View(func(tx *bolt.Tx) error {
		manifest := tx.Bucket(manifestBucketName)
		c := manifest.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			ps := ProjSnapManifest{}
			if err := json.Unmarshal(v, &ps); err != nil {
				return err
			}
			psm.meta.ManifestSnapshots[string(k)] = ps
		}
		return nil
	})
}

func (psm *ProjSnapMaster) dumpProjSnapshot(snapName string, appSnapshots []AppSnapshot) (seq uint64, err error) {
	oldSnap, _ := psm.meta.ManifestSnapshots[snapName]
	err = psm.db.Update(func(tx *bolt.Tx) error {
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
		manifest := tx.Bucket(manifestBucketName)
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

func (psm *ProjSnapMaster) RegisterApplication(name string, app apps.AppPacker) {
	psm.specPackers[strings.ToLower(name)] = app
}

func (psm *ProjSnapMaster) GetPacker(appName string) apps.AppPacker {
	appName = strings.ToLower(appName)
	if packer, ok := psm.specPackers[appName]; ok {
		return packer
	}
	return psm.generalPacker
}

func (psm *ProjSnapMaster) quitAllApplication(appNames map[string]struct{}) {
	hasTerm := false
	for app := range appNames {
		if app != "iTerm2" {
			log.Printf("quit %s, err: %v", app, psm.GetPacker(app).Quit(app))
		} else {
			hasTerm = true
		}
	}
	// todo: hard code
	if hasTerm {
		_ = psm.GetPacker("iTerm2").Quit("iTerm2")
	}
}

func (psm *ProjSnapMaster) SaveSnapshot(snapName string) (bool, error) {
	appNames, err := psm.getAllApplication()
	if err != nil {
		return false, err
	}
	if err := psm.wm.TakeSnapshot(); err != nil {
		return false, err
	}

	appSnapshots := make([]AppSnapshot, 0)
	for app := range appNames {
		conf, err := psm.GetPacker(app).Pack(psm.opt.configDir, app)
		if err != nil {
			return false, fmt.Errorf("%s occur fail, err: %v", app, err)
		}
		// todo: save要关联正常，restore关联也要正常，现在是随机
		for i := range conf {
			wind, _ := psm.wm.GetWindowInfo(app) // ignore error
			appSnapshots = append(appSnapshots, AppSnapshot{AppConfig: &conf[i], WindowInfo: wind})
		}
	}

	ctxID, err := psm.dumpProjSnapshot(snapName, appSnapshots)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("SaveWorkSpace Success, ctxID: %d, alias: %s", ctxID, snapName)

	if psm.opt.quit {
		psm.quitAllApplication(appNames)
	}
	return true, nil
}

func (psm *ProjSnapMaster) loadSnapshot(snapName string) []AppSnapshot {
	// alias
	snapshot, ok := psm.meta.ManifestSnapshots[snapName]
	if !ok {
		log.Fatal("no found snapName")
	}
	appSnapshots := make([]AppSnapshot, 0)

	_ = psm.db.View(func(tx *bolt.Tx) error {
		snap := tx.Bucket(SnapshotsBucketName)
		data := snap.Get([]byte(snapshot.SnapshotKey))
		return json.Unmarshal(data, &appSnapshots)
	})
	return appSnapshots
}

func (psm *ProjSnapMaster) openAppFromSnapshot(appSnapshots []AppSnapshot, realRunning map[string]struct{}) error {
	for i, conf := range appSnapshots {
		log.Printf("[%d/%d] Opening %s, args: %v\n", i+1, len(appSnapshots), conf.AppName, conf.Args)
		_, running := realRunning[conf.AppName]
		if err := psm.GetPacker(conf.AppName).Unpack(conf.AppConfig, running); err != nil {
			return err
		}
	}
	return nil
}

func (psm *ProjSnapMaster) SwitchSnapshot(snapName string) error {
	appSnapshots := psm.loadSnapshot(snapName)
	realRunning, err := psm.getAllApplication()
	if err != nil {
		return err
	}
	// open all
	if err := psm.openAppFromSnapshot(appSnapshots, realRunning); err != nil {
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
			_ = psm.GetPacker(app).Quit(app)
		}
	}
	// wait
	time.Sleep(3 * time.Second)
	// get current opened windows
	if err := psm.wm.TakeSnapshot(); err != nil {
		return err
	}
	// restore windows
	for _, conf := range appSnapshots {
		_ = psm.wm.RestoreWindow(conf.WindowInfo)
	}
	return nil
}

func (psm *ProjSnapMaster) RestoreSnapshot(snapName string) error {
	appSnapshots := psm.loadSnapshot(snapName)
	// open app, ignore current whether is opened
	if err := psm.openAppFromSnapshot(appSnapshots, map[string]struct{}{}); err != nil {
		return err
	}
	// wait app
	time.Sleep(3 * time.Second)
	// get current opened windows
	if err := psm.wm.TakeSnapshot(); err != nil {
		return err
	}
	// restore windows
	for _, conf := range appSnapshots {
		_ = psm.wm.RestoreWindow(conf.WindowInfo)
	}

	return nil
}

func (psm *ProjSnapMaster) getAllApplication() (map[string]struct{}, error) {
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
