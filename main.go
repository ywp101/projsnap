package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"projsnap/utils"
	"time"
)

/*
启用“辅助功能权限”
System Events 访问其他 app 的窗口信息需要你授权：
打开：系统设置 -> 隐私与安全 -> 辅助功能
启用当前运行脚本的终端（如 iTerm、Terminal、Script Editor、你的 Go 程序等）
*/

var configDir, _ = utils.ExpandUser("~/.projsnap/")
var quitFlag bool
var snapName string
var rmIndex int

var rootCmd = &cobra.Command{
	Use:   "projsnap",
	Short: "Save the current snapshot and restore it when needed",
	Long:  "Save the current snapshot and restore it when needed",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use `projsnap snapshot` or `projsnap restore` to start.")
	},
}

var snapshotCmd = &cobra.Command{
	Use:     "take_snap",
	Aliases: []string{"take"},
	Short:   "Save the current snapshot",
	Run: func(cmd *cobra.Command, args []string) {
		if snapName == "" {
			log.Println("You should input snapName(--name [snapshot] or -n [snapshot])")
			return
		}
		ws := NewWorkspace(&ProjSnapOptions{
			quit:      quitFlag,
			configDir: configDir,
		})
		if err := ws.Open(); err != nil {
			log.Fatal(err)
		}
		defer ws.Close()
		if ok, err := ws.SaveSnapshot(snapName); !ok || err != nil {
			log.Printf("SaveSnapshot fail, ok: %v, err: %v\n", ok, err)
		}
	},
}

var switchCmd = &cobra.Command{
	Use:     "switch",
	Aliases: []string{"sw"},
	Short:   "switch specific snapshot",
	Run: func(cmd *cobra.Command, args []string) {
		if snapName == "" {
			log.Println("You should input snapName(--name [snapshot] or -n [snapshot])")
			return
		}
		ws := NewWorkspace(&ProjSnapOptions{
			configDir: configDir,
		})
		if err := ws.Open(); err != nil {
			log.Fatal(err)
		}
		defer ws.Close()
		if err := ws.SwitchSnapshot(snapName); err != nil {
			log.Printf("SwitchSnapshot occur error: %v\n", err)
		}
	},
}

var restoreCmd = &cobra.Command{
	Use:     "restore",
	Aliases: []string{"rs"},
	Short:   "restore specific snapshot(use after reboot)",
	Run: func(cmd *cobra.Command, args []string) {
		if snapName == "" {
			log.Println("You should input snapName(--name [snapshot] or -n [snapshot])")
			return
		}
		ws := NewWorkspace(&ProjSnapOptions{
			configDir: configDir,
		})
		if err := ws.Open(); err != nil {
			log.Fatal(err)
		}
		defer ws.Close()
		if err := ws.RestoreSnapshot(snapName); err != nil {
			log.Printf("RestoreSnapshot occur error: %v\n", err)
		}
	},
}

var listSnapshotCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "ll"},
	Short:   "list ManifestSnapshots",
	Run: func(cmd *cobra.Command, args []string) {
		ws := NewWorkspace(&ProjSnapOptions{
			configDir: configDir,
		})
		if err := ws.Open(); err != nil {
			log.Fatal(err)
		}
		defer ws.Close()
		i := 1
		for _, snapshot := range ws.ListSnapshots() {
			if snapshot.SnapshotName == snapshot.SnapshotKey {
				fmt.Printf("[%d] %s\t%s\n", i, snapshot.SnapshotKey, time.Unix(snapshot.Ctime, 0).String())
			} else {
				fmt.Printf("[%d] %s(%s)\t%s\n", i, snapshot.SnapshotName, snapshot.SnapshotKey, time.Unix(snapshot.Ctime, 0).String())
			}
			i++
		}
		if i == 1 {
			fmt.Println("no found any ManifestSnapshots.")
		}
	},
}

var rmSnapshotCmd = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"rm"},
	Short:   "remove ManifestSnapshots",
	Run: func(cmd *cobra.Command, args []string) {
		if snapName == "" {
			log.Println("You should input snapName(--name [snapshot] or -n [snapshot])")
			return
		}
		ws := NewWorkspace(&ProjSnapOptions{
			configDir: configDir,
		})
		if err := ws.Open(); err != nil {
			log.Fatal(err)
		}
		defer ws.Close()
		if err := ws.RemoveSnapshots(snapName); err != nil {
			fmt.Printf("remove snapshots fail, err:%v\n", err)
		} else {
			fmt.Println("remove snapshots success!")
		}
	},
}

func init() {
	snapshotCmd.Flags().BoolVarP(&quitFlag, "quit", "q", false, "Exit when saving snapshot")
	snapshotCmd.Flags().StringVarP(&snapName, "name", "n", "", "snapshot name")
	switchCmd.Flags().StringVarP(&snapName, "name", "n", "", "snapshot name")
	restoreCmd.Flags().StringVarP(&snapName, "name", "n", "", "snapshot name")
	rmSnapshotCmd.Flags().StringVarP(&snapName, "name", "n", "", "snapshot name")
	rootCmd.AddCommand(snapshotCmd, restoreCmd, listSnapshotCmd, rmSnapshotCmd, switchCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
