package ndb_api

type DatabaseProvisionRequest struct {
	DatabaseType             string           `json:"databaseType"`
	Name                     string           `json:"name"`
	DatabaseDescription      string           `json:"databaseDescription"`
	SoftwareProfileId        string           `json:"softwareProfileId"`
	SoftwareProfileVersionId string           `json:"softwareProfileVersionId"`
	ComputeProfileId         string           `json:"computeProfileId"`
	NetworkProfileId         string           `json:"networkProfileId"`
	DbParameterProfileId     string           `json:"dbParameterProfileId"`
	NewDbServerTimeZone      string           `json:"newDbServerTimeZone"`
	CreateDbServer           bool             `json:"createDbserver"`
	NodeCount                int              `json:"nodeCount"`
	NxClusterId              string           `json:"nxClusterId"`
	SSHPublicKey             string           `json:"sshPublicKey"`
	Clustered                bool             `json:"clustered"`
	AutoTuneStagingDrive     bool             `json:"autoTuneStagingDrive"`
	TimeMachineInfo          TimeMachineInfo  `json:"timeMachineInfo"`
	ActionArguments          []ActionArgument `json:"actionArguments"`
	Nodes                    []Node           `json:"nodes"`
}

type DatabaseDeprovisionRequest struct {
	Delete               bool `json:"delete"`
	Remove               bool `json:"remove"`
	SoftRemove           bool `json:"softRemove"`
	Forced               bool `json:"forced"`
	DeleteTimeMachine    bool `json:"deleteTimeMachine"`
	DeleteLogicalCluster bool `json:"deleteLogicalCluster"`
}

type DatabaseServerDeprovisionRequest struct {
	Delete            bool `json:"delete"`
	Remove            bool `json:"remove"`
	SoftRemove        bool `json:"softRemove"`
	DeleteVgs         bool `json:"deleteVgs"`
	DeleteVmSnapshots bool `json:"deleteVmSnapshots"`
}
