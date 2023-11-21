package container

import (
	"encoding/json"
	"fmt"
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
	dir := DefaultInfoLocation + "/" + name
	err := os.MkdirAll(dir, 0622)
	if err != nil {
		return "", util.CleanFnNil, err
	}
	configFile := dir + "/" + ConfigName
	file, err := os.Create(configFile)
	if err != nil {
		return "", util.CleanFnNil, err
	}
	defer file.Close()
	bytes, _ := json.Marshal(info)
	_, err = file.Write(bytes)
	return name, func() error {
		logrus.Debugf("clean container(%s) info", name)
		return os.RemoveAll(dir)
	}, err
}

func GetContainerInfo(cName string) (ContainerInfo, error) {
	exist := ContainerExist(cName)
	var info ContainerInfo
	if !exist {
		return info, fmt.Errorf("container %s isn't exist", cName)
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
