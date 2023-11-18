package util

import (
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func CreateCId() string {
	return "c" + strconv.FormatInt(time.Now().UnixMilli(), 36)
}

func NewDir(dirPath string, perm os.FileMode) (cleanFn CleanFn, err error) {
	err = os.MkdirAll(dirPath, perm)
	cleanFn = func() error {
		logrus.Debug("remove dir " + dirPath)
		return os.RemoveAll(dirPath)
	}
	return
}

func BindOutput(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
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
