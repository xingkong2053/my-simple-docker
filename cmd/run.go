/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/pkg/errors"
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
	parent, writePipe, err := NewParentProcess(tty)
	if err != nil {
		logrus.Error("new parent process error " + err.Error())
		return
	}
	defer func() {
		err := DeleteWorkSpace("/root/", "/root/mnt/")
		if err != nil {
			logrus.Error(errors.Wrap(err, "delete workspace").Error())
		}
	}()

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
	_ = writePipe.Close()
	parent.Wait()
}

func NewParentProcess(tty bool) (*exec.Cmd, *os.File, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	command := exec.Command("/proc/self/exe", "init")
	// execute command with namespace
	command.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}

	// 把readPipe发送给子进程
	command.ExtraFiles = []*os.File{r}
	mntUrl := "/root/mnt/"
	err = NewWorkSpace("/root/", mntUrl)
	if err != nil {
		return nil, nil, err
	}
	command.Dir = mntUrl
	if tty {
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}
	return command, w, nil
}

func NewWorkSpace(rootUrl string, mntUrl string) error {
	err := CreateReadonlyLayer(rootUrl)
	if err != nil {
		return errors.Wrap(err, "create readonly layer")
	}
	err = CreateWriteLayer(rootUrl)
	if err != nil {
		return errors.Wrap(err, "create write layer")
	}
	err = CreateMountPoint(rootUrl, mntUrl)
	return errors.Wrap(err, "create mount point")
}

func CreateReadonlyLayer(rootUrl string) error {
	// 将busybox.tar解压到busybox目录下
	bbDir := rootUrl + "busybox/"
	exist, err := PathExists(bbDir)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}
	err = os.Mkdir(bbDir, 0777)
	if err != nil {
		return err
	}
	_, err = exec.Command("tar", "-xvf", rootUrl+"busybox.tar", "-C", bbDir).CombinedOutput()
	return err
}

func PathExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

func CreateWriteLayer(rootUrl string) error {
	// 创建writeLayer文件夹作为容器的唯一可写层
	return os.Mkdir(rootUrl+"writeLayer/", 0777)
}

func CreateMountPoint(rootUrl string, mntUrl string) error {
	err := os.Mkdir(mntUrl, 0777)
	if err != nil {
		return err
	}
	err = os.Mkdir(rootUrl + "workdir/", 0777)
	if err != nil {
		return err
	}
	// https://askubuntu.com/questions/109413/how-do-i-use-overlayfs
	dirs := fmt.Sprintf("upperdir=%swriteLayer,lowerdir=%sbusybox,workdir=%sworkdir", rootUrl, rootUrl, rootUrl)
	logrus.Infof("exec command: mount -t overlay -o %s none %s", dirs, mntUrl)
	cmd := exec.Command("mount", "-t", "overlay" /* ubuntu 不再支持aufs, 使用overlay*/, "-o", dirs, "none", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func DeleteWorkSpace(rootUrl string, mntUrl string) error {
	// unmount & delete mnt
	cmd := exec.Command("umount", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	err = os.RemoveAll(mntUrl)
	if err != nil {
		return err
	}
	err = os.RemoveAll(rootUrl + "workdir")
	if err != nil {
		return err
	}
	// delete write layer
	return os.RemoveAll(rootUrl + "writeLayer/")
}
