package ndb_api

type SnapshotRequest struct {
	Name              string            `json:"name"`
	SnapshotLcmConfig SnapshotLcmConfig `json:"lcmConfig"`
}

type SnapshotLcmConfig struct {
	SnapshotLCMConfigDetailed SnapshotLcmConfigDetailed `json:"snapshotLCMConfig"`
}

type SnapshotLcmConfigDetailed struct {
	ExpiryDetails ExpiryDetails `json:"expiryDetails"`
}
