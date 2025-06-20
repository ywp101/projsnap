package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/twmb/murmur3"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
)

func Hash(s string) string {
	hasher := murmur3.New32()
	_, _ = hasher.Write([]byte(s))
	return strconv.Itoa(int(hasher.Sum32()))
}

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
	return slices.DeleteFunc(strings.Split(string(b), "\n"), func(s string) bool { return s == "" }), nil
}

func RecoverBakFile(dst string, srcData string) error {
	srcFd, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer srcFd.Close()
	_, err = srcFd.WriteString(srcData)
	return err
}

func BakFile(dstDir, src string) (string, error) {
	srcFd, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer srcFd.Close()
	dst := fmt.Sprintf("%s/%s.%d", dstDir, filepath.Base(src), time.Now().Unix())
	dstFd, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer dstFd.Close()
	_, err = io.Copy(dstFd, srcFd)
	return dst, err
}

func ExpandUser(path string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	jetbrainHome := "$USER_HOME$"
	shellHome := "~"
	if strings.HasPrefix(path, shellHome) {
		return filepath.Join(home, path[len(shellHome):]), nil
	}
	if strings.HasPrefix(path, jetbrainHome) {
		return filepath.Join(home, path[len(jetbrainHome):]), nil
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

func Uint64ToBytes(n uint64) []byte {
	buf := make([]byte, 8) // uint64 占 8 字节
	binary.BigEndian.PutUint64(buf, n)
	return buf
}

func GetPIDFromAppName(appName string) ([]int, error) {
	cmd := exec.Command("pgrep", appName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ps failed: %w", err)
	}
	pids := make([]int, 0)
	for _, v := range strings.Split(string(output), "\n") {
		if v == "" {
			continue
		}
		pid, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("parse pid failed: %w", err)
		}
		pids = append(pids, pid)
	}
	sort.Ints(pids)
	return pids, nil
}
