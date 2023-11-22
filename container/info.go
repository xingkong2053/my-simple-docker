package container

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"mydocker/util"
	"os"
	"path"
	"strconv"
	"time"
)

func CreateContainerInfo(id string, pid int, cmd string, name string) (string, util.CleanFn, error) {
	info := ContainerInfo{
		Pid:    strconv.Itoa(pid),
		Id:     id,
		Name:   name,
		Cmd:    cmd,
		Create: time.Now().Format("2006-01-02 15:04:05"),
		Status: Running,
	}
	err := SetContainerInfo(info, true)
	return name, func() error {
		logrus.Debugf("clean container(%s) info", name)
		return os.RemoveAll(path.Join(DefaultInfoLocation, name))
	}, err
}

func GetContainerInfo(cName string) (ContainerInfo, error) {
	exist := ContainerExist(cName)
	var info ContainerInfo
	if !exist {
		return info, fmt.Errorf("container %s doesn't exist", cName)
	}
	filePath := path.Join(DefaultInfoLocation, cName, ConfigName)
	f, err := os.Open(filePath)
	if err != nil {
		return info, err
	}
	bytes, err := io.ReadAll(f)
	if err != nil {
		return info, err
	}
	err = json.Unmarshal(bytes, &info)
	return info, err
}

func SetContainerInfo(info ContainerInfo, create bool) error {
	cName := info.Name
	if cName == "" {
		return errors.New("set container info error: name is empty")
	}
	dirPath := path.Join(DefaultInfoLocation, cName)
	exists, err := util.PathExists(dirPath)
	if err != nil {
		return err
	}
	if !exists {
		if create {
			err := os.MkdirAll(dirPath, 0622)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("directory %s doesn't exist", dirPath)
		}
	}
	filePath := path.Join(dirPath, ConfigName)
	exists, err = util.PathExists(filePath)
	if err != nil {
		return err
	}
	if !exists {
		if create {
			f, err := os.Create(filePath)
			if err != nil {
				return err
			}
			_ = f.Close()
		} else {
			return fmt.Errorf("file %s doesn't exist", filePath)
		}
	}
	bytes, _ := json.Marshal(info)
	return os.WriteFile(filePath, bytes, 0622)
}

func RemoveContainer(cName string) error {
	info, err := GetContainerInfo(cName)
	if err != nil {
		return err
	}
	if info.Status != Stop {
		return errors.New("couldn't remove running container")
	}
	return os.RemoveAll(path.Join(DefaultInfoLocation, cName))
}
