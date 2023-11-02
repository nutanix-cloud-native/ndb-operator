package ndb_api

type SnapshotResponse struct {
	Id                      string                   `json:"id"`
	Name                    string                   `json:"name"`
	Description             string                   `json:"description"`
	OwnerId                 string                   `json:"ownerId"`
	DateCreated             string                   `json:"dateCreated"`
	DateModified            string                   `json:"dateModified"`
	Properties              []interface{}            `json:"properties"`
	Tags                    []interface{}            `json:"tags"`
	SnapshotId              string                   `json:"snapshotId"`
	SnapshotUuid            string                   `json:"snapshotUuid"`
	ProtectionDomainId      string                   `json:"protectionDomainId"`
	TimeMachineId           string                   `json:"timeMachineId"`
	DatabaseNodeId          string                   `json:"databaseNodeId"`
	AppInfoVersion          string                   `json:"appInfoVersion"`
	Status                  string                   `json:"status"`
	Type                    string                   `json:"type"`
	ApplicableTypes         []string                 `json:"applicableTypes"`
	SnapshotTimeStamp       string                   `json:"snapshotTimeStamp"`
	Info                    Info                     `json:"info"`
	Metadata                SnapshotResponseMetadata `json:"metadata"`
	TimeZone                string                   `json:"timeZone"`
	SnapshotSize            float32                  `json:"snapshotSize"`
	Processed               bool                     `json:"processed"`
	DatabaseSnapshot        bool                     `json:"databaseSnapshot"`
	FromTimeStamp           string                   `json:"fromTimeStamp"`
	ToTimeStamp             string                   `json:"toTimeStamp"`
	DbserverId              string                   `json:"dbserverId"`
	DbserverName            string                   `json:"dbserververName"`
	DbserverIp              string                   `json:"dbServerIp"`
	LcmConfig               interface{}              `json:"lcmConfig"`
	TypeFrequency           string                   `json:"typeFrequency"`
	ApplicableTypeFrequency string                   `json:"applicableTypeFrequency"`
}

type Info struct {
	SecureInfo interface{} `json:"secureInfo"`
	Info       interface{} `json:"info"`
}

type SnapshotResponseMetadata struct {
	SecureInfo     interface{} `json:"secureInfo"`
	Info           interface{} `json:"info"`
	DeregisterInfo interface{} `json:"deregisterInfo"`
	FromTimeStamp  string      `json:"fromTimeStamp"`
	ToTimeStamp    string      `json:"toTimeStamp"`
}
