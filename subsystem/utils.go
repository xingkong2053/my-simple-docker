package subsystem

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

// GetAbsoluteCgroupPath 查找subsystem挂载的hierarchy相对路径对应的cgroup在虚拟文件中的路径
func GetAbsoluteCgroupPath(systemName string, cgroupPath string, create bool) (string, error) {
	cgroupRoot, err := FindCgroupMountPoint(systemName)
	if err != nil {
		return "", err
	}
	absolute := path.Join(cgroupRoot, cgroupPath)
	_, err = os.Stat(absolute)
	if err == nil {
		return absolute, nil
	}

	// has error
	if os.IsNotExist(err) {
		if create {
			err := os.Mkdir(absolute, 0755)
			if err != nil {
				return "", err
			}
			return absolute, nil
		} else {
			return "", fmt.Errorf("cgroup %s don't exist", cgroupPath)
		}
	}
	return "", err
}

// FindCgroupMountPoint 通过 /proc/self/mountinfo文件, 寻找cgroup根目录
func FindCgroupMountPoint(subsystem string) (string, error) {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", err
	}
	defer f.Close()
	// mountinfo文件行结构
	// 32 23 0:27 / /sys/fs/cgroup/memory rw,nosuid,nodev,noexec,relatime shared:9 - cgroup2 cgroup2 rw,memory
	scanner := bufio.NewScanner(f)
	var lineNo int
	for scanner.Scan() {
		line := scanner.Text()
		lineNo++
		if strings.HasSuffix(line, ",memory") {
			fields := strings.Split(line, " ")
			if len(fields) < 5 {
				errStr := "/proc/self/mountinfo line:%d: %s doesn't contain cgroup root path"
				return "", fmt.Errorf(errStr, lineNo, line)
			}
			return fields[4], nil
		}
	}
	return "", scanner.Err()
}

func ApplyProcessToCgroup(subsystem string, cgroupPath string, pid int) error {
	absoluteCgroupPath, err := GetAbsoluteCgroupPath(subsystem, cgroupPath, false)
	if err != nil {
		return err
	}
	taskFilePath := path.Join(absoluteCgroupPath, "tasks")
	bytes := []byte(strconv.Itoa(pid))
	return os.WriteFile(taskFilePath, bytes, 0644)
}

func RemoveCgroup(subsystem string, cgroupPath string) error {
	absoluteCgroupPath, err := GetAbsoluteCgroupPath(subsystem, cgroupPath, false)
	if err != nil {
		return err
	}
	return os.Remove(absoluteCgroupPath)
}
