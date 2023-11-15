package subsystem

import (
	"os"
	"path"
)

type MemorySubsystem struct{}

func (m *MemorySubsystem) Name() string {
	return "memory"
}

// 设置cgroupPath对应的cgroup的内存资源限制
func (m *MemorySubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	if res.MemoryLimit != "" {
		return nil
	}
	subsysCgroupPath, err := GetAbsoluteCgroupPath(m.Name(), cgroupPath, true)
	if err != nil {
		return err
	}
	memLimitP := path.Join(subsysCgroupPath, "memory.limit_in_bytes")
	return os.WriteFile(memLimitP, []byte(res.MemoryLimit), 0644)
}

// 将一个进程加入到cgroupPath对应的cgroup中
func (m *MemorySubsystem) Apply(cgroupPath string, pid int) error {
	return ApplyProcessToCgroup(m.Name(), cgroupPath, pid)
}

// 删除cgroupPath对应的cgroup
func (m *MemorySubsystem) Remove(cgroupPath string) error {
	return RemoveCgroup(m.Name(), cgroupPath)
}
