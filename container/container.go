package container

import (
	"mydocker/util"
	"path"

	"github.com/sirupsen/logrus"
)

const (
	DefaultInfoLocation = "/var/run/mydocker"
	ConfigName          = "config.json"
	LogFile             = "container.log"
)

func ContainerExist(cName string) bool {
	exist, err := util.PathExists(path.Join(DefaultInfoLocation, cName))
	if err != nil {
		logrus.Warn("check path exist error:", err.Error())
		return false
	}
	return exist
}