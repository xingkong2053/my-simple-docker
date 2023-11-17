/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os/exec"
)

// commitCmd represents the commit command
var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "A brief description of your command",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		Commit(args[0])
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
}

func Commit(imageName string) {
	_, err := exec.Command("tar", "-czf", "/root/"+imageName+".tar",
		"-C", "/root/mnt", ".").CombinedOutput()
	if err != nil {
		logrus.Error(err.Error())
	}
}
