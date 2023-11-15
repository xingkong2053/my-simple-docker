package subsystem

import (
	"os"
	"path"
)

type CpuSubsystem struct{}

func (m *CpuSubsystem) Name() string {
	return "cpu"
}

// 设置cgroupPath对应的cgroup的内存资源限制
func (m *CpuSubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	if res.CpuShare != "" {
		return nil
	}
	subsysCgroupPath, err := GetAbsoluteCgroupPath(m.Name(), cgroupPath, true)
	if err != nil {
		return err
	}
	memLimitP := path.Join(subsysCgroupPath, "cpu.shares")
	return os.WriteFile(memLimitP, []byte(res.CpuShare), 0644)
}

// 将一个进程加入到cgroupPath对应的cgroup中
func (m *CpuSubsystem) Apply(cgroupPath string, pid int) error {
	return ApplyProcessToCgroup(m.Name(), cgroupPath, pid)
}

// 删除cgroupPath对应的cgroup
func (m *CpuSubsystem) Remove(cgroupPath string) error {
	return RemoveCgroup(m.Name(), cgroupPath)
}
