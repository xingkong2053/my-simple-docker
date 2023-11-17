package container

type ContainerStatus uint

const (
	Running ContainerStatus = iota
	Stop
	Exit
)

type ContainerInfo struct {
	Pid    string          `json:"pid"`
	Id     string          `json:"id"`
	Name   string          `json:"name"`
	Cmd    string          `json:"cmd"`
	Create string          `json:"create"`
	Status ContainerStatus `json:"status"`
}
