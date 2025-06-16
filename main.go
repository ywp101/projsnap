package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"workspace/utils"
)

var configDir, _ = utils.ExpandUser("~/.projctx/")
var quitFlag bool
var ctxVersion string
var aliasName string

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
		ws := NewWorkspace(&WorkspaceOptions{
			quit:      quitFlag,
			configDir: configDir,
		})
		if ok, err := ws.SaveWorkspace(aliasName); !ok || err != nil {
			log.Printf("SaveWorkspace fail, ok: %v, err: %v\n", ok, err)
		}
	},
}

var restoreCmd = &cobra.Command{
	Use:     "restore",
	Aliases: []string{"rs"},
	Short:   "restore specific snapshot",
	Run: func(cmd *cobra.Command, args []string) {
		if ctxVersion == "" {
			log.Printf("not found ctxVersion: %s\n", ctxVersion)
		}
		ws := NewWorkspace(&WorkspaceOptions{
			configDir: configDir,
		})
		if err := ws.LoadWorkspace(ctxVersion); err != nil {
			log.Printf("LoadWorkspace occur error: %v\n", err)
		}
	},
}

var listSnapshotCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "list Snapshots",
	Run: func(cmd *cobra.Command, args []string) {
		ws := NewWorkspace(&WorkspaceOptions{
			configDir: configDir,
		})
		i := 1
		for alias, ctxID := range ws.ListSnapshots() {
			if alias == ctxID {
				fmt.Printf("[%d] %s\n", i, alias)
			} else {
				fmt.Printf("[%d] %s(%s)", i, alias, ctxID)
			}
			i++
		}
		if i == 1 {
			fmt.Println("no found any Snapshots.")
		}
	},
}

func init() {
	snapshotCmd.Flags().BoolVarP(&quitFlag, "quit", "q", false, "Exit when saving snapshot")
	snapshotCmd.Flags().StringVarP(&aliasName, "alias", "a", "", "Snapshot alias name")
	restoreCmd.Flags().StringVarP(&ctxVersion, "ctx", "c", "", "ctx version")
	rootCmd.AddCommand(snapshotCmd, restoreCmd, listSnapshotCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
