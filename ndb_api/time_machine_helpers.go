package ndb_api

func GenerateSnapshotRequest(name string, expiryDateTimezone string, ExpireInDays string) *SnapshotRequest {
	return &SnapshotRequest{
		Name: name,
		SnapshotLcmConfig: SnapshotLcmConfig{
			SnapshotLCMConfigDetailed: SnapshotLcmConfigDetailed{
				ExpiryDetails: ExpiryDetails{
					ExpiryDateTimezone: expiryDateTimezone,
					ExpireInDays:       ExpireInDays,
				},
			},
		},
	}
}
