package utils

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

type CheckRunningAppFn func() ([]string, error)

func OpenMultiAppByRetry(fn CheckRunningAppFn, appName string, args ...string) error {
	retryArgs := args
	retryCnt := 5
	for len(retryArgs) != 0 && retryCnt > 0 {
		err := OpenMultiApp(appName, retryArgs...)
		if err != nil {
			return err
		}
		retryArgs, err = fn()
		retryCnt--
		if err != nil {
			return err
		}
	}
	return nil
}

func OpenMultiApp(appName string, args ...string) error {
	log.Printf("open multi app: %v", args)
	timeout := 5 * time.Second
	for _, arg := range args {
		if err := OpenApp(appName, arg); err != nil {
			return err
		}
		time.Sleep(timeout)
	}
	time.Sleep(timeout)
	return nil
}

func OpenApp(appName string, args ...string) error {
	allArgs := make([]string, 0)
	if appName != "" {
		allArgs = append(allArgs, "-a", appName)
	}
	allArgs = append(allArgs, args...)

	cmd := exec.Command("open", allArgs...)
	return cmd.Run()
}

func RunOsascript(script string) ([]string, error) {
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	items := make([]string, 0)
	for _, item := range strings.Split(string(out), ",") {
		items = append(items, strings.TrimSpace(item))
	}
	return items, nil
}

func GracefulQuit(appName string) error {
	script := fmt.Sprintf(`if application "%s" is running then quit app "%s"`, appName, appName)
	if err := exec.Command("osascript", "-e", script).Run(); err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	return nil
}
