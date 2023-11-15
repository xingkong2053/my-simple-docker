package subsystem

import (
	"os"
	"path"
)

type CpusetSubsystem struct{}

func (m *CpusetSubsystem) Name() string {
	return "cpuset"
}

// 设置cgroupPath对应的cgroup的内存资源限制
func (m *CpusetSubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	if res.CpuSet != "" {
		return nil
	}
	subsysCgroupPath, err := GetAbsoluteCgroupPath(m.Name(), cgroupPath, true)
	if err != nil {
		return err
	}
	memLimitP := path.Join(subsysCgroupPath, "cpuset.cpus")
	return os.WriteFile(memLimitP, []byte(res.MemoryLimit), 0644)
}

// 将一个进程加入到cgroupPath对应的cgroup中
func (m *CpusetSubsystem) Apply(cgroupPath string, pid int) error {
	return ApplyProcessToCgroup(m.Name(), cgroupPath, pid)
}

// 删除cgroupPath对应的cgroup
func (m *CpusetSubsystem) Remove(cgroupPath string) error {
	return RemoveCgroup(m.Name(), cgroupPath)
}
