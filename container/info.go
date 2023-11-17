package container

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"mydocker/util"
	"os"
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
