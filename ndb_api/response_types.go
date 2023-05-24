package ndb_api

type DatabaseResponse struct {
	Id            string         `json:"id"`
	Name          string         `json:"name"`
	Status        string         `json:"status"`
	DatabaseNodes []DatabaseNode `json:"databaseNodes"`
	Properties    []Property     `json:"properties"`
}

type TaskInfoSummaryResponse struct {
	Name                 string                    `json:"name"`
	WorkId               string                    `json:"workId"`
	OperationId          string                    `json:"operationId"`
	DbServerId           string                    `json:"dbserverId"`
	Message              string                    `json:"messgae"`
	EntityId             string                    `json:"entityId"`
	EntityName           string                    `json:"entityName"`
	EntityType           string                    `json:"entityType"`
	Status               string                    `json:"status"`
	AssociatedOperations []TaskInfoSummaryResponse `json:"associatedOperations"`
	DependencyReport     interface{}               `json:"dependencyReport"`
}

type ProfileResponse struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	EngineType      string `json:"engineType"`
	LatestVersionId string `json:"latestVersionId"`
	Topology        string `json:"topology"`
	SystemProfile   bool   `json:"systemProfile"`
	Status          string `json:"status"`
}

type SLAResponse struct {
	Id                 string `json:"id"`
	Name               string `json:"name"`
	UniqueName         string `json:"uniqueName"`
	Description        string `json:"description"`
	DailyRetention     int    `json:"dailyRetention"`
	WeeklyRetention    int    `json:"weeklyRetention"`
	MonthlyRetention   int    `json:"monthlyRetention"`
	QuarterlyRetention int    `json:"quarterlyRetention"`
	YearlyRetention    int    `json:"yearlyRetention"`
}
