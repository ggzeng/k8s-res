// Package cmd
// Copyright © 2022 Zeng Ganghui <zengganghui@gmail.com>
package cmd

import (
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print app version",
	//Long: ``
	Run: func(cmd *cobra.Command, args []string) {
		// don't need to print version
		// utils.PrintFullVersion
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
