package subsystem

import "github.com/sirupsen/logrus"

type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string
	CpuSet      string
}

type Subsystem interface {
	Name() string
	// 设置某个cgroup在这个Subsystem中的资源限制
	Set(cgroupPath string, res *ResourceConfig) error
	// 将进程添加到某个cgroup中
	Apply(cgroupPath string, pid int) error
	// 移除某个cgroup
	Remove(cgroupPath string) error
}

var SubsystemSet = []Subsystem{
	&MemorySubsystem{},
	&CpuSubsystem{},
	&CpusetSubsystem{},
}

type CgroupManager struct {
	// 创建的cgroup目录相对于各root cgroup目录的路径
	Path     string
	Resource *ResourceConfig
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{Path: path}
}

func (c *CgroupManager) Set(res *ResourceConfig) error {
	for _, subsystem := range SubsystemSet {
		err := subsystem.Set(c.Path, res)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CgroupManager) Apply(pid int) error {
	for _, subsystem := range SubsystemSet {
		err := subsystem.Apply(c.Path, pid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CgroupManager) Destroy() {
	for _, subsystem := range SubsystemSet {
		err := subsystem.Remove(c.Path)
		if err != nil {
			logrus.Warn(err.Error())
		}
	}
}
