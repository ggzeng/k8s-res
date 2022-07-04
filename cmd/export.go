// Package cmd
// Copyright © 2022 Zeng Ganghui <zengganghui@gmail.com>
package cmd

import (
	"github.com/spf13/cobra"
	k8client "k8res/internal/k8s/client"
	"k8res/internal/process"
	//"k8res/internal/process"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "export current pods resource",
	//Long: ``
	Run: exportStart,
}

func exportStart(cmd *cobra.Command, args []string) {
	k8 := k8client.New("")
	store := make(process.AllPodResStore)
	if err := process.GetPodRes(k8, store); err != nil {
		panic(err)
	}
	process.ExportPodRes(store)
}

func init() {
	rootCmd.AddCommand(exportCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// exportCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// exportCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
