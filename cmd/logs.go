/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"mydocker/container"
	"os"
	"path"

	"github.com/spf13/cobra"
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Print logs of a container",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := ReadLogs(args[0])
		if err != nil {
			logrus.Errorf("read %s logs error: %s", args[0], err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
}

func ReadLogs(cName string) error {
	filePath := path.Join(container.DefaultInfoLocation, cName, container.LogFile)
	logrus.Info("open log file: " + filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(os.Stdout, string(bytes))
	return err
}
