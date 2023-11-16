/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pkg/errors"

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
	err = setupMount()
	if err != nil {
		return err
	}

	path, err := exec.LookPath(cmdArr[0])
	if err != nil {
		return errors.Wrap(err, "exec look path")
	}

	logrus.Info("Find path " + path)
	// replace init process with user command
	return errors.Wrap(syscall.Exec(path, cmdArr[0:], os.Environ()), "exec command "+path)
}

func setupMount() error {
	// https://github.com/xianlubird/mydocker/issues/41
	// systemd 加入linux之后, mount namespace 就变成 shared by default, 所以你必须显式
	// 声明你要这个新的mount namespace独立。
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")

	pwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "get wd")
	}
	logrus.Info("current location is " + pwd)
	err = pivotRoot(pwd)
	if err != nil {
		return errors.Wrap(err, "pivot root")
	}

	mountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	// 挂在proc文件系统，方便使用ps等命令
	syscall.Mount("proc", "/proc", "proc", uintptr(mountFlags), "")

	// tmpfs基于内存的文件系统
	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
	return errors.Wrap(nil, "mount proc")
}

func readUserCommand() ([]string, error) {
	pipe := os.NewFile(uintptr(3 /*文件描述符*/), "pipe")
	bytes, err := io.ReadAll(pipe) // 阻塞
	if err != nil {
		return nil, errors.Wrap(err, "read user command")
	}
	logrus.Info("init command invoked. arg is " + string(bytes))
	return strings.Split(string(bytes), " "), nil
}

func pivotRoot(root string /* /root/busybox */) error {
	// 为了使老root和新root不在同一个文件系统下，需要重新mount
	err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, "")
	if err != nil {
		return errors.Wrap(err, "mount root")
	}
	pivotDir := filepath.Join(root, ".pivot_root") // /root/busybox/.pivot_root
	_, err = os.Stat(pivotDir)
	if err != nil {
		if !os.IsExist(err) {
			err := os.Mkdir(pivotDir, 0777)
			if err != nil {
				return err
			}
		}
	}

	err = syscall.PivotRoot(root, pivotDir)
	if err != nil {
		return errors.Wrap(err, "syscall pivot_root")
	}
	err = syscall.Chdir("/")
	if err != nil {
		return errors.Wrap(err, "chdir /")
	}
	// 切换到新文件系统后pivotDir变为 /.pivot_root, 实际上和之前是同一个文件夹
	pivotDir = filepath.Join("/", ".pivot_root")	
	err = syscall.Unmount(pivotDir, syscall.MNT_DETACH)
	if err != nil {
		return errors.Wrap(err, "unmount")
	}
	return os.Remove(pivotDir)
}
