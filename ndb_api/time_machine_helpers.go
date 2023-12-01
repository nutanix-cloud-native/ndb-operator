package ndb_api

func GenerateSnapshotRequest(name string, expiryDateTimezone string, ExpireInDays int) *SnapshotRequest {
	return &SnapshotRequest{
		Name: name,
		SnapshotLcmConfig: SnapshotLcmConfig{
			SnapshotLCMConfigDetailed: SnapshotLcmConfigDetailed{
				ExpiryDetails: SnapshotExpiryDetails{
					ExpiryDateTimezone: expiryDateTimezone,
					ExpireInDays:       ExpireInDays,
				},
			},
		},
	}
}
