package ndb_api

type SnapshotResponse struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	SnapshotId    string `json:"snapshotId"`
	SnapshotUuid  string `json:"snapshotUuid"`
	TimeMachineId string `json:"timeMachineId"`
}

type AllSnapshotResponse struct {
	ID                 string        `json:"id"`
	Name               string        `json:"name"`
	Description        interface{}   `json:"description"`
	OwnerID            string        `json:"ownerId"`
	DateCreated        string        `json:"dateCreated"`
	DateModified       string        `json:"dateModified"`
	AccessLevel        interface{}   `json:"accessLevel"`
	Properties         []interface{} `json:"properties"`
	Tags               []interface{} `json:"tags"`
	SnapshotID         string        `json:"snapshotId"`
	SnapshotUUID       string        `json:"snapshotUuid"`
	NxClusterID        string        `json:"nxClusterId"`
	ProtectionDomainID string        `json:"protectionDomainId"`
	ParentSnapshotID   interface{}   `json:"parentSnapshotId"`
	TimeMachineID      string        `json:"timeMachineId"`
	DatabaseNodeID     string        `json:"databaseNodeId"`
	AppInfoVersion     string        `json:"appInfoVersion"`
	Status             string        `json:"status"`
	Type               string        `json:"type"`
	ApplicableTypes    []string      `json:"applicableTypes"`
	SnapshotTimeStamp  string        `json:"snapshotTimeStamp"`
	Info               struct {
		SecureInfo      interface{} `json:"secureInfo"`
		Info            interface{} `json:"info"`
		LinkedDatabases []struct {
			ID           string `json:"id"`
			DatabaseName string `json:"databaseName"`
			Status       string `json:"status"`
			Info         struct {
				Info struct {
					CreatedBy string `json:"created_by"`
				} `json:"info"`
			} `json:"info"`
			AppConsistent bool        `json:"appConsistent"`
			Message       interface{} `json:"message"`
			Clone         bool        `json:"clone"`
		} `json:"linkedDatabases"`
		Databases          interface{} `json:"databases"`
		DatabaseGroupID    interface{} `json:"databaseGroupId"`
		MissingDatabases   interface{} `json:"missingDatabases"`
		ReplicationHistory interface{} `json:"replicationHistory"`
	} `json:"info"`
	Metadata struct {
		SecureInfo                           interface{}   `json:"secureInfo"`
		Info                                 interface{}   `json:"info"`
		DeregisterInfo                       interface{}   `json:"deregisterInfo"`
		FromTimeStamp                        string        `json:"fromTimeStamp"`
		ToTimeStamp                          string        `json:"toTimeStamp"`
		ReplicationRetryCount                int           `json:"replicationRetryCount"`
		LastReplicationRetryTimestamp        interface{}   `json:"lastReplicationRetryTimestamp"`
		LastReplicationRetrySourceSnapshotID interface{}   `json:"lastReplicationRetrySourceSnapshotId"`
		Async                                bool          `json:"async"`
		Standby                              bool          `json:"standby"`
		CurationRetryCount                   int           `json:"curationRetryCount"`
		OperationsUsingSnapshot              []interface{} `json:"operationsUsingSnapshot"`
	} `json:"metadata"`
	Metric struct {
		LastUpdatedTimeInUTC interface{} `json:"lastUpdatedTimeInUTC"`
		Storage              struct {
			LastUpdatedTimeInUTC        interface{} `json:"lastUpdatedTimeInUTC"`
			ControllerNumIops           interface{} `json:"controllerNumIops"`
			ControllerAvgIoLatencyUsecs interface{} `json:"controllerAvgIoLatencyUsecs"`
			Size                        int         `json:"size"`
			AllocatedSize               int         `json:"allocatedSize"`
			UsedSize                    int         `json:"usedSize"`
			Unit                        string      `json:"unit"`
		} `json:"storage"`
	} `json:"metric"`
	SoftwareSnapshotID             string      `json:"softwareSnapshotId"`
	SoftwareDatabaseSnapshot       bool        `json:"softwareDatabaseSnapshot"`
	DbServerStorageMetadataVersion int         `json:"dbServerStorageMetadataVersion"`
	Sanitised                      bool        `json:"sanitised"`
	SanitisedFromSnapshotID        interface{} `json:"sanitisedFromSnapshotId"`
	TimeZone                       string      `json:"timeZone"`
	Processed                      bool        `json:"processed"`
	DatabaseSnapshot               bool        `json:"databaseSnapshot"`
	FromTimeStamp                  string      `json:"fromTimeStamp"`
	ToTimeStamp                    string      `json:"toTimeStamp"`
	DbserverID                     interface{} `json:"dbserverId"`
	DbserverName                   interface{} `json:"dbserverName"`
	DbserverIP                     interface{} `json:"dbserverIp"`
	ReplicatedSnapshots            interface{} `json:"replicatedSnapshots"`
	SoftwareSnapshot               interface{} `json:"softwareSnapshot"`
	SanitisedSnapshots             interface{} `json:"sanitisedSnapshots"`
	SnapshotFamily                 interface{} `json:"snapshotFamily"`
	SnapshotTimeStampDate          int64       `json:"snapshotTimeStampDate"`
	LcmConfig                      interface{} `json:"lcmConfig"`
	SnapshotSize                   int         `json:"snapshotSize"`
	ParentSnapshot                 bool        `json:"parentSnapshot"`
}

type LcmConfig struct {
	ExpiryDetails struct {
		RemindBeforeInDays int    `json:"remindBeforeInDays"`
		EffectiveTimestamp string `json:"effectiveTimestamp"`
		ExpiryTimestamp    string `json:"expiryTimestamp"`
		ExpiryDateTimezone string `json:"expiryDateTimezone"`
		UserCreated        bool   `json:"userCreated"`
		ExpireInDays       int    `json:"expireInDays"`
	} `json:"expiryDetails"`
	RefreshDetails struct {
		RefreshInDays       int    `json:"refreshInDays"`
		RefreshInHours      int    `json:"refreshInHours"`
		RefreshInMonths     int    `json:"refreshInMonths"`
		LastRefreshDate     string `json:"lastRefreshDate"`
		NextRefreshDate     string `json:"nextRefreshDate"`
		RefreshTime         string `json:"refreshTime"`
		RefreshDateTimezone string `json:"refreshDateTimezone"`
	} `json:"refreshDetails"`
	PreDeleteCommand struct {
		Command string `json:"command"`
	} `json:"preDeleteCommand"`
	PostDeleteCommand struct {
		Command string `json:"command"`
	} `json:"postDeleteCommand"`
}
