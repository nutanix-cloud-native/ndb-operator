package ndb_api

type DatabaseNode struct {
	Id               string `json:"id"`
	Name             string `json:"name"`
	DatabaseServerId string `json:"dbServerId"`
}

type TimeMachineInfo struct {
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	SlaId            string            `json:"slaId"`
	Schedule         map[string]string `json:"schedule"`
	Tags             []string          `json:"tags"`
	AutoTuneLogDrive bool              `json:"autoTuneLogDrive"`
}

type ActionArgument struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Node struct {
	Properties []string `json:"properties"`
	VmName     string   `json:"vmName"`
}

type Property struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
}
