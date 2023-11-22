/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/sirupsen/logrus"
	"mydocker/container"

	"github.com/spf13/cobra"
)

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "remove a stopped container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := container.RemoveContainer(args[0])
		if err != nil {
			logrus.Error(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
