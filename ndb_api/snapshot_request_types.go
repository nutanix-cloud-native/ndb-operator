package ndb_api

type SnapshotRequest struct {
	Name              string            `json:"name"`
	SnapshotLcmConfig SnapshotLcmConfig `json:"lcmConfig"`
}

type SnapshotLcmConfig struct {
	SnapshotLCMConfigDetailed SnapshotLcmConfigDetailed `json:"snapshotLCMConfig"`
}

type SnapshotLcmConfigDetailed struct {
	ExpiryDetails SnapshotExpiryDetails `json:"expiryDetails"`
}

type SnapshotExpiryDetails struct {
	ExpiryDateTimezone string `json:"expiryDateTimezone"`
	ExpireInDays       int    `json:"expireInDays"`
}
