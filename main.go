package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"projctx/utils"
	"time"
)

/*
启用“辅助功能权限”
System Events 访问其他 app 的窗口信息需要你授权：
打开：系统设置 -> 隐私与安全 -> 辅助功能
启用当前运行脚本的终端（如 iTerm、Terminal、Script Editor、你的 Go 程序等）
*/

var configDir, _ = utils.ExpandUser("~/.projctx/")
var quitFlag bool
var ctxVersion string
var aliasName string
var rmIndex int

var rootCmd = &cobra.Command{
	Use:   "projctx",
	Short: "Save the current snapshot and restore it when needed",
	Long:  "Save the current snapshot and restore it when needed",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use `projctx snapshot` or `projctx restore` to start.")
	},
}

var snapshotCmd = &cobra.Command{
	Use:     "snapshot",
	Aliases: []string{"ss"},
	Short:   "Save the current snapshot",
	Run: func(cmd *cobra.Command, args []string) {
		ws := NewWorkspace(&ProjectCtxOptions{
			quit:      quitFlag,
			configDir: configDir,
		})
		defer ws.Close()
		if ok, err := ws.SaveWorkspace(aliasName); !ok || err != nil {
			log.Printf("SaveWorkspace fail, ok: %v, err: %v\n", ok, err)
		}
	},
}

var switchCmd = &cobra.Command{
	Use:     "switch",
	Aliases: []string{"sw"},
	Short:   "switch specific snapshot",
	Run: func(cmd *cobra.Command, args []string) {
		if ctxVersion == "" {
			log.Printf("not found ctxVersion: %s\n", ctxVersion)
			return
		}
		ws := NewWorkspace(&ProjectCtxOptions{
			configDir: configDir,
		})
		defer ws.Close()
		if err := ws.SwitchWorkspace(ctxVersion); err != nil {
			log.Printf("SwitchWorkspace occur error: %v\n", err)
		}
	},
}

var restoreCmd = &cobra.Command{
	Use:     "restore",
	Aliases: []string{"rs"},
	Short:   "restore specific snapshot(use after reboot)",
	Run: func(cmd *cobra.Command, args []string) {
		if ctxVersion == "" {
			log.Printf("not found ctxVersion: %s\n", ctxVersion)
			return
		}
		ws := NewWorkspace(&ProjectCtxOptions{
			configDir: configDir,
		})
		defer ws.Close()
		if err := ws.LoadWorkspace(ctxVersion); err != nil {
			log.Printf("LoadWorkspace occur error: %v\n", err)
		}
	},
}

var listSnapshotCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "ll"},
	Short:   "list Snapshots",
	Run: func(cmd *cobra.Command, args []string) {
		ws := NewWorkspace(&ProjectCtxOptions{
			configDir: configDir,
		})
		defer ws.Close()
		i := 1
		for _, snapshot := range ws.ListSnapshots() {
			if snapshot.SnapshotAlias == snapshot.SnapshotKey {
				fmt.Printf("[%d] %s\t%s\n", i, snapshot.SnapshotKey, time.Unix(snapshot.Ctime, 0).String())
			} else {
				fmt.Printf("[%d] %s(%s)\t%s\n", i, snapshot.SnapshotAlias, snapshot.SnapshotKey, time.Unix(snapshot.Ctime, 0).String())
			}
			i++
		}
		if i == 1 {
			fmt.Println("no found any Snapshots.")
		}
	},
}

var rmSnapshotCmd = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"rm"},
	Short:   "remove Snapshots",
	Run: func(cmd *cobra.Command, args []string) {
		ws := NewWorkspace(&ProjectCtxOptions{
			configDir: configDir,
		})
		defer ws.Close()
		if err := ws.RemoveSnapshots(ctxVersion); err != nil {
			fmt.Printf("remove snapshots fail, err:%v\n", err)
		} else {
			fmt.Println("remove snapshots success!")
		}
	},
}

func init() {
	snapshotCmd.Flags().BoolVarP(&quitFlag, "quit", "q", false, "Exit when saving snapshot")
	snapshotCmd.Flags().StringVarP(&aliasName, "alias", "a", "", "Snapshot alias name")
	switchCmd.Flags().StringVarP(&ctxVersion, "ctx", "c", "", "ctx version")
	restoreCmd.Flags().StringVarP(&ctxVersion, "ctx", "c", "", "ctx version")
	rmSnapshotCmd.Flags().StringVarP(&ctxVersion, "ctx", "c", "", "ctx version")
	rootCmd.AddCommand(snapshotCmd, restoreCmd, listSnapshotCmd, rmSnapshotCmd, switchCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
