/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"mydocker/container"
	"strconv"
	"syscall"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop a container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := StopContainer(args[0])
		if err != nil {
			logrus.Error(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func StopContainer(cName string) error {
	info, err := container.GetContainerInfo(cName)
	if err != nil {
		return err
	}
	pid, err := strconv.Atoi(info.Pid)
	if err != nil {
		return err
	}
	err = syscall.Kill(pid, syscall.SIGTERM)
	if err != nil {
		return err
	}
	// 修改容器状态
	info.Pid = ""
	info.Status = container.Stop
	return container.SetContainerInfo(info, false)
}
