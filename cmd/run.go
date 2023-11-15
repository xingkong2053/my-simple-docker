/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"mydocker/subsystem"
	"os"
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var tty bool
var resource subsystem.ResourceConfig

// runCmd represents the run command
// 退出之后执行 `sudo mount -t proc proc /proc`
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run container",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		cmd := args[0]
		logrus.Info("start run " + cmd)
		Run(cmd, tty)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVarP(&tty, "tty", "t", false, "enable tty")
	runCmd.Flags().StringVarP(&resource.MemoryLimit, "memory", "m", "", "memory limit")
	runCmd.Flags().StringVar(&resource.CpuShare, "cpushare", "", "cpu share limit")
	runCmd.Flags().StringVar(&resource.CpuSet, "cpuset", "", "cpu set limit")
}

func Run(cmd string, tty bool) {
	parent, writePipe, err := NewParentProcess(tty, cmd)
	if err != nil {
		logrus.Error("new parent process error " + err.Error())
		return
	}
	// Start() 会clone出来一个namespace隔离的进程
	// 然后在子进程中，调用/proc/self/exe(./mydocker)
	if err := parent.Start(); err != nil {
		logrus.Error(err.Error())
		return
	}

	// 创建cgroup manager并设置资源限制
	manager := subsystem.NewCgroupManager("mydocker-cgroup")
	defer manager.Destroy()
	// 设置资源限制
	err = manager.Set(&resource)
	if err != nil {
		logrus.Error(err.Error())
		return
	}
	err = manager.Apply(parent.Process.Pid)
	if err != nil {
		logrus.Error(err.Error())
		return
	}

	_, err = writePipe.WriteString(cmd)
	if err != nil {
		logrus.Error("send cmd to child process error: " + err.Error())
		return
	}
	parent.Wait()
}

func NewParentProcess(tty bool, cmd string) (*exec.Cmd, *os.File, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	command := exec.Command("/proc/self/exe", "init", cmd)
	// execute command with namespace
	command.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}

	// 把readPipe发送给子进程
	command.ExtraFiles = []*os.File{r}
	if tty {
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}
	return command, w, nil
}
