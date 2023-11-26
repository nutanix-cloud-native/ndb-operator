package ndb_api

type SnapshotResponse struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	SnapshotId    string `json:"snapshotId"`
	SnapshotUuid  string `json:"snapshotUuid"`
	TimeMachineId string `json:"timeMachineId"`
}
