package container

type ContainerStatus string

const (
	Running ContainerStatus = "running"
	Stop                    = "stop"
	Exit                    = "exit"
)

type ContainerInfo struct {
	Pid    string          `json:"pid"`
	Id     string          `json:"id"`
	Name   string          `json:"name"`
	Cmd    string          `json:"cmd"`
	Create string          `json:"create"`
	Status ContainerStatus `json:"status"`
}
