/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Init container process run user's process in container.",
	RunE: func(_ *cobra.Command, args []string) error {
		cmd := args[0]
		logrus.Info("init command invoked. arg is " + cmd)
		return RunContainerInit(cmd, nil)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func RunContainerInit(command string, args []string) error {
	// var err error
	mountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV

	// 挂在proc文件系统，方便使用ps等命令
	err := syscall.Mount("proc", "/proc", "proc", uintptr(mountFlags), "")
	if err != nil {
		return err
	}
	// replace init process with user command
	return syscall.Exec(command, []string{command}, os.Environ())
}
