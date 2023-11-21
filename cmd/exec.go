/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "exec a command into container",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ExecContainer(args[0], args[1:])
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
}

func ExecContainer(cName string, commands []string) {
	logrus.Infof("container name: %s, commands: %s", cName, strings.Join(commands, " "))
}
