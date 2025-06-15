package utils

import (
	"encoding/json"
	"errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"log"
	"strings"
)

type RecentFile struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func ReadRecentFiles(dbPath string) ([]string, error) {
	dbPath, err := ExpandUser(dbPath)
	if err != nil {
		return nil, err
	}
	db, err := leveldb.OpenFile(dbPath, &opt.Options{
		ReadOnly: true,
	})
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	iter := db.NewIterator(nil, nil)
	defer iter.Release()
	for iter.Next() {
		key := string(iter.Key())
		if strings.Contains(key, "recent") {
			value, err := decodeUTF16LE(iter.Value())
			if err != nil {
				return nil, err
			}
			files := make([]RecentFile, 0)
			if err = json.Unmarshal([]byte(value), &files); err != nil {
				return nil, err
			}
			filePaths := make([]string, 0)
			for _, file := range files {
				filePaths = append(filePaths, file.ID)
			}
			return filePaths, err
		}
	}
	return nil, errors.New("not found")
}
