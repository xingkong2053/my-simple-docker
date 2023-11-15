/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Init container process run user's process in container.",
	RunE: func(_ *cobra.Command, args []string) error {
		return RunContainerInit()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func RunContainerInit() (err error) {
	cmdArr, err := readUserCommand() // 阻塞，只有当参数读取完之后才mount
	if err != nil {
		return errors.Wrap(err, "run init command")
	}
	if len(cmdArr) == 0 {
		return errors.New("user command is empty")
	}
	path, err := exec.LookPath(cmdArr[0])
	if err != nil {
		return errors.Wrap(err, "exec look path")
	}

	err = mountProc()
	if err != nil {
		return err
	}

	logrus.Info("Find path " + path)
	// replace init process with user command
	return errors.Wrap(syscall.Exec(path, cmdArr[0:], os.Environ()), "exec command "+path)
}

func mountProc() error {
	mountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	// 挂在proc文件系统，方便使用ps等命令
	return errors.Wrap(syscall.Mount("proc", "/proc", "proc", uintptr(mountFlags), ""), "mount proc")
}

func readUserCommand() ([]string, error) {
	pipe := os.NewFile(uintptr(3 /*文件描述符*/), "pipe")
	bytes, err := io.ReadAll(pipe) // 阻塞
	if err != nil {
		c
		return nil, errors.Wrap(err, "read user command")
	}
	logrus.Info("init command invoked. arg is " + string(bytes))
	return strings.Split(string(bytes), " "), nil
}
