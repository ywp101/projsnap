package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"testing"
)

func TestNewWorkspacePack(t *testing.T) {
	ws := NewWorkspace(&ProjSnapOptions{
		configDir: configDir,
	})
	if err := ws.Open(); err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	_, _ = ws.GetPacker("Microsoft Edge").Pack("", "Microsoft Edge")
}

func TestNewWorkspaceUnPack(t *testing.T) {
	ws := NewWorkspace(&ProjSnapOptions{
		configDir: configDir,
	})
	if err := ws.Open(); err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	appName := "Microsoft Edge"
	appSnapshots := ws.loadSnapshot("fuck")
	for _, sshot := range appSnapshots {
		if sshot.AppName == appName {
			_ = ws.GetPacker(appName).Unpack(sshot.AppConfig, true)
		}
	}
}

func TestNewWorkspacelistAll(t *testing.T) {
	ws := NewWorkspace(&ProjSnapOptions{
		configDir: configDir,
	})
	if err := ws.Open(); err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	_ = ws.db.Update(func(tx *bolt.Tx) error {
		ssIter := tx.Bucket([]byte("snapshots")).Cursor()
		ctxIter := tx.Bucket([]byte("ctx")).Cursor()
		for k, v := ssIter.First(); k != nil; k, v = ssIter.Next() {
			fmt.Println(string(k), string(v))
		}
		fmt.Println("-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=")
		for k, v := ctxIter.First(); k != nil; k, v = ctxIter.Next() {
			fmt.Println(string(k), string(v))
		}
		return nil
	})
}
