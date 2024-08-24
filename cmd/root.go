// Package cmd
// Copyright © 2022 Zeng Ganghui <zengganghui@gmail.com>
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8res/pkg/config"
	log "k8res/pkg/logger"
	"k8res/pkg/utils"
)

var (
	runMode    string
	logLevel   string
	namespaces []string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k8res",
	Short: "Get k8s pods cpu and mem resource usage",
	Long:  `Get all k8s pods cpu and mem request, limit and current usage with a special namespace.`,
	Args:  cobra.MinimumNArgs(1),
	//Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	utils.PrintFullVersion()
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&runMode, "mode", "m", "prod", "run mode with: prod, dev, test")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log", "info", "logger lever: debug, info, warn, error ")
	if err := viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log")); err != nil {
		fmt.Printf("FATAIL: %s", err)
		os.Exit(1)
	}
	rootCmd.PersistentFlags().StringArrayVarP(&namespaces, "namespaces", "n", []string{"all"}, "use special namespace list or all")
	if err := viper.BindPFlag("app.namespaces", rootCmd.PersistentFlags().Lookup("namespaces")); err != nil {
		fmt.Printf("FATAIL: %s", err)
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if err := config.ViperInit(runMode, "k8res"); err != nil {
		fmt.Printf("ERROR: init config failed, %v", err)
	}
	log.Initialize() // need following config init
	if err := config.Save(); err != nil {
		fmt.Printf("ERROR: save curr config failed, %v", err)
	}
}
