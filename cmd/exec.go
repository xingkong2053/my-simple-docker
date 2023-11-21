/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"mydocker/container"
	_ "mydocker/nsenter"
	"os"
	"os/exec"
	"strings"
)

const (
	EnvExecPid = "mydocker_pid"
	EnvExecCmd = "mydocker_cmd"
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "exec a command into container",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		err := ExecContainer(args[0], args[1:])
		if err != nil {
			logrus.Error(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
}

func ExecContainer(cName string, commands []string) error {
	cmdS := strings.Join(commands, " ")
	info, err := container.GetContainerInfo(cName)
	if err != nil {
		return errors.Wrap(err, "get container info")
	}
	logrus.Infof("container(name: %s, pid: %s),commands: %s", cName, info.Pid, cmdS)

	err = os.Setenv(EnvExecCmd, cmdS)
	if err != nil {
		return err
	}
	err = os.Setenv(EnvExecPid, info.Pid)
	if err != nil {
		return err
	}

	// 指定环境变量后再次执行exec命令
	cmd := exec.Command("/proc/self/exe", "exec" /* mydocker exec */)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return errors.Wrap(cmd.Run(), "run command")
}
