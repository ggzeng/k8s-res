// Package cmd
// Copyright © 2022 Zeng Ganghui <zengganghui@gmail.com>
package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	k8client "k8res/internal/k8s/client"
	"k8res/internal/process"
	"k8res/pkg/logger"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	interval int32
)

// monitorCmd represents the monitor command
var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "monitor pods resource, stop with Ctl+C",
	//Long: ``
	Run: monitorStart,
}

func monitorStart(cmd *cobra.Command, args []string) {
	k8 := k8client.New("")
	store := make(process.AllPodResStore)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan interface{})
	go func(done chan<- interface{}, sigCh <-chan os.Signal) {
		for {
			select {
			case <-sigCh:
				fmt.Println()
				fmt.Println("monitor stopped")
				done <- true
				return
			default:
				if err := process.GetPodRes(k8, store); err != nil {
					logger.Error(err)
				}
				fmt.Print(".")
				time.Sleep(time.Duration(interval) * time.Second)
			}
		}
	}(done, sigCh)
	<-done
	fmt.Println("EXPORT DATA:")
	process.ExportPodRes(store)
}

func init() {
	rootCmd.AddCommand(monitorCmd)

	monitorCmd.Flags().Int32VarP(&interval, "interval", "i", 10, "monitor interval seconds")
	if err := viper.BindPFlag("app.interval", monitorCmd.Flags().Lookup("interval")); err != nil {
		fmt.Printf("FATAIL: %s", err)
		os.Exit(1)
	}

}
