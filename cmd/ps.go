/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"mydocker/container"
	"os"
	"path"
	"text/tabwriter"
)

// psCmd represents the ps command
var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "list docker containers",
	Run: func(cmd *cobra.Command, args []string) {
		ListContainers()
	},
}

func init() {
	rootCmd.AddCommand(psCmd)
}

func ListContainers() {
	entries, err := os.ReadDir(container.DefaultInfoLocation)
	if err != nil {
		logrus.Error("read dir error", err.Error())
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		filePath := path.Join(container.DefaultInfoLocation, entry.Name(), container.ConfigName)
		f, err := os.Open(filePath)
		if err != nil {
			logrus.Error("open file error", err.Error())
			continue
		}
		bytes, err := io.ReadAll(f)
		if err != nil {
			logrus.Error("read all error", err.Error())
			continue
		}
		var info container.ContainerInfo
		err = json.Unmarshal(bytes, &info)
		if err != nil {
			logrus.Error(err.Error())
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t\n",
			info.Id, info.Name, info.Pid, info.Status, info.Cmd, info.Create)
	}
	err = w.Flush()
	if err != nil {
		logrus.Error("flush error")
	}
}
