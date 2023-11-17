/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"mydocker/container"
	"mydocker/subsystem"
	"mydocker/util"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
)

var cleanFns []util.CleanFn

func CollectCleanFn(fn util.CleanFn, err error) error {
	if err != nil {
		return err
	}
	cleanFns = append(cleanFns, fn)
	return nil
}

var cleanup = func() error {
	if len(cleanFns) == 0 {
		return nil
	}
	logrus.Info("do cleanup")
	i := len(cleanFns) - 1
	for i >= 0 {
		fn := cleanFns[i]
		if fn == nil {
			continue
		}
		err := fn()
		if err != nil {
			return err
		}
		i--
	}
	return nil
}

var tty bool
var detach bool
var resource subsystem.ResourceConfig
var volume string
var name string

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
	runCmd.Flags().BoolVarP(&detach, "detach", "d", false, "detach")
	runCmd.MarkFlagsMutuallyExclusive("tty", "detach")
	runCmd.Flags().StringVarP(&resource.MemoryLimit, "memory", "m", "", "memory limit")
	runCmd.Flags().StringVar(&resource.CpuShare, "cpushare", "", "cpu share limit")
	runCmd.Flags().StringVar(&resource.CpuSet, "cpuset", "", "cpu set limit")
	runCmd.Flags().StringVarP(&volume, "volume", "v", "", "mount volume")
	runCmd.Flags().StringVarP(&name, "name", "n", "", "container name")
}

func Run(cmd string, tty bool) {
	id := util.CreateCId()
	if name == "" {
		name = id
	}
	parent, writePipe, err, clean := NewParentProcess(tty, name)
	if err != nil {
		logrus.Error("new parent process error " + err.Error())
		return
	}
	defer func() {
		if !tty || clean == nil {
			return
		}
		err := clean()
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

	// 记录容器信息
	_, cleanInfo, err := container.CreateContainerInfo(id, parent.Process.Pid, cmd, name)
	if err != nil {
		logrus.Error(errors.Wrap(err, "create container info").Error())
		return
	}
	defer func() {
		if !tty {
			return
		}
		err := cleanInfo()
		if err != nil {
			logrus.Error(errors.Wrap(err, "clean container info").Error())
		}
	}()

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
	if tty {
		_ = parent.Wait()
	}
}

func NewParentProcess(tty bool, containerName string) (*exec.Cmd, *os.File, error, util.CleanFn) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, err, nil
	}
	command := exec.Command("/proc/self/exe", "init")
	// execute command with namespace
	command.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}

	// 把readPipe发送给子进程
	command.ExtraFiles = []*os.File{r}
	mntUrl := path.Join("/root/mnt", containerName)
	cleanup, err := NewWorkSpace(path.Join("/root", containerName), mntUrl)
	if err != nil {
		return nil, nil, err, cleanup
	}
	command.Dir = mntUrl
	if tty {
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}
	return command, w, nil, cleanup
}

func NewWorkSpace(rootUrl string, mntUrl string) (util.CleanFn, error) {
	// readonly layer
	err := CreateReadonlyLayer(rootUrl)
	if err != nil {
		return cleanup, errors.Wrap(err, "create readonly layer")
	}
	// write layer
	err = CollectCleanFn(CreateWriteLayer(rootUrl))
	if err != nil {
		return cleanup, errors.Wrap(err, "create write layer")
	}
	// workdir
	err = CollectCleanFn(util.NewDir(path.Join(rootUrl, "workdir"), 0777))
	if err != nil {
		return cleanup, err
	}
	// create mnt dir
	err = CollectCleanFn(util.NewDir(mntUrl, 0777))
	if err != nil {
		return cleanup, err
	}
	// mount mnt
	err = CollectCleanFn(Mount(rootUrl, mntUrl))
	if err != nil {
		return cleanup, err
	}
	// mount volume
	if volume != "" {
		src, dist, err := parseVolume(volume)
		if err != nil {
			return cleanup, err
		}
		logrus.Debugf("mount volume %s to %s", src, dist)
		exists, err := util.PathExists(src)
		if err != nil {
			return cleanup, err
		}
		if !exists {
			logrus.Debugf("path %s don't exist, create", src)
			err := os.Mkdir(src, 0777)
			if err != nil {
				return cleanup, err
			}
		}
		err = CollectCleanFn(util.NewDir(mntUrl+dist, 0777))
		if err != nil {
			return cleanup, err
		}
		// mount volume
		err = CollectCleanFn(MountDist(src, mntUrl+dist))
		if err != nil {
			return cleanup, err
		}
	}
	return cleanup, nil
}

func parseVolume(volume string) (string, string, error) {
	arr := strings.Split(volume, ":")
	err := errors.New("volume option value: " + volume + " is not correct")
	if len(arr) < 2 {
		return "", "", err
	}
	src, dist := arr[0], arr[1]
	if src == "" || dist == "" {
		return "", "", err
	}
	return src, dist, nil
}

func MountDist(src, dist string) (util.CleanFn, error) {
	cmd := exec.Command("mount", "--bind", src, dist)
	logrus.Debug(cmd.String())
	util.BindOutput(cmd)
	return func() error {
		umount := exec.Command("umount", dist)
		logrus.Debug(umount.String())
		util.BindOutput(umount)
		return umount.Run()
	}, cmd.Run()
}

func CreateReadonlyLayer(rootUrl string) error {
	// 将busybox.tar解压到busybox目录下
	bbDir := rootUrl + "/busybox"
	exist, err := util.PathExists(bbDir)
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
	_, err = exec.Command("tar", "-xvf", rootUrl+"/busybox.tar", "-C", bbDir).CombinedOutput()
	return err
}

func CreateWriteLayer(rootUrl string) (util.CleanFn, error) {
	// 创建writeLayer文件夹作为容器的唯一可写层
	return util.NewDir(rootUrl+"/writeLayer", 0777)
}

func Mount(rootUrl string, mntUrl string) (util.CleanFn, error) {
	// https://askubuntu.com/questions/109413/how-do-i-use-overlayfs
	option := fmt.Sprintf("upperdir=%s/writeLayer,lowerdir=%s/busybox,workdir=%s/workdir", rootUrl, rootUrl, rootUrl)
	logrus.Infof("exec command: mount -t overlay -o %s none %s", option, mntUrl)
	cmd := exec.Command("mount", "-t", "overlay" /* ubuntu 不再支持aufs, 使用overlay*/, "-o", option, "none", mntUrl)
	unMount := func() error {
		logrus.Debug("unmount mnt")
		cmd := exec.Command("umount", mntUrl)
		util.BindOutput(cmd)
		return cmd.Run()
	}
	util.BindOutput(cmd)
	err := cmd.Run()
	return unMount, err
}
