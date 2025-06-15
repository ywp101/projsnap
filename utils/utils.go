package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf16"
)

func ReadFileToStringList(file string) ([]string, error) {
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	b, err := io.ReadAll(fd)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(b), "\n"), nil
}

func RecoverBakFile(bakFile string) error {
	bakFd, err := os.Open(bakFile)
	if err != nil {
		return err
	}
	defer bakFd.Close()
	srcFile := bakFile[:strings.LastIndex(bakFile, ".")]
	srcFd, err := os.Create(srcFile)
	if err != nil {
		return err
	}
	defer srcFd.Close()
	_, err = io.Copy(srcFd, bakFd)
	return err
}

func BakFile(src string) (string, error) {
	srcFd, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer srcFd.Close()
	dst := fmt.Sprintf("%s.%d", src, time.Now().Unix())
	dstFd, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer dstFd.Close()
	_, err = io.Copy(dstFd, srcFd)
	return dst, err
}

func ExpandUser(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[1:]), nil
	}
	return path, nil
}

func decodeUTF16LE(b []byte) (string, error) {
	if len(b)%2 != 0 {
		b = b[1:]
	}

	u16s := make([]uint16, 0, len(b)/2)
	buf := bytes.NewReader(b)

	for {
		var u uint16
		err := binary.Read(buf, binary.LittleEndian, &u)
		if err != nil {
			break
		}
		u16s = append(u16s, u)
	}
	runes := utf16.Decode(u16s)
	return string(runes), nil
}
