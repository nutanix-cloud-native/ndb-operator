package ndb_api

type SnapshotRequest struct {
	Name              string            `json:"name"`
	SnapshotLcmConfig SnapshotLcmConfig `json:"lcmConfig"`
	TimeMachineId     string            `json:'timemachine_id'`
}

type SnapshotLcmConfig struct {
	SnapshotLCMConfigDetailed SnapshotLcmConfigDetailed `json:"snapshotLCMConfig"`
}

type SnapshotLcmConfigDetailed struct {
	ExpiryDetails ExpiryDetails `json:"expiryDetails"`
}
