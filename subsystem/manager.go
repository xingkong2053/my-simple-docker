package subsystem

type ResourceConfig struct {
	MemoryLimit string
	CpuShare string
	CpuSet string
}

type Subsystem interface {
	// 设置某个cgroup在这个Subsystem中的资源限制
	Set(path string, res *ResourceConfig) error
	// 将进程添加到某个cgroup中
	Apply(path string, pid int) error
	// 移除某个cgroup
	Remove(path string) error
}

type CgroupManager struct {
	// 创建的cgroup目录相对于各root cgroup目录的路径
	Path string
	Resource *ResourceConfig
}

