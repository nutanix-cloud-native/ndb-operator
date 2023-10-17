/*
Copyright 2022-2023 Nutanix, Inc.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ndb_api

import (
	"context"

	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// This function generates and returns a request for cloning a database on NDB
func GenerateCloningRequest(ctx context.Context, ndb_client *ndb_client.NDBClient, database DatabaseInterface, reqData map[string]interface{}) (requestBody *DatabaseCloneRequest, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.GenerateCloningRequest", "database name", database.GetName())

	sourceDatabase, _ := GetDatabaseById(ctx, ndb_client, database.GetCloneSourceDBId())
	databaseType := GetDatabaseTypeFromEngine(sourceDatabase.Type)
	// Fetch the required profiles for the database
	profilesMap, err := ResolveProfiles(ctx, ndb_client, databaseType, database.GetProfileResolvers())
	if err != nil {
		log.Error(err, "Error occurred while getting required profiles", "database name", database.GetName(), "isClone", database.IsClone())
		return
	}
	// Required for dbParameterProfileIdInstance in MSSQL action args
	reqData[common.PROFILE_MAP_PARAM] = profilesMap

	// Creating a provisioning request based on the database type
	requestBody = &DatabaseCloneRequest{
		Name:           database.GetName(),
		Description:    database.GetDescription(),
		CreateDbServer: true,
		Clustered:      false,
		NxClusterId:    database.GetClusterId(),
		// SSHPublicKey populated by request appenders for non mssql dbs
		DbServerId:               "",
		DbServerClusterId:        "",
		DbserverLogicalClusterId: "",
		TimeMachineId:            sourceDatabase.TimeMachineId,
		SnapshotId:               database.GetCloneSnapshotId(),
		UserPitrTimestamp:        "",
		TimeZone:                 database.GetTimeZone(),
		LatestSnapshot:           false,
		NodeCount:                1,
		Nodes: []Node{
			{
				VmName:              database.GetName() + "_CLONE_VM",
				ComputeProfileId:    profilesMap[common.PROFILE_TYPE_COMPUTE].Id,
				NetworkProfileId:    profilesMap[common.PROFILE_TYPE_NETWORK].Id,
				NewDbServerTimeZone: "",
				NxClusterId:         database.GetClusterId(),
				Properties:          make([]string, 0),
			},
		},
		// Added by request appenders as per the engine
		ActionArguments:            []ActionArgument{},
		Tags:                       make([]interface{}, 0),
		VmPassword:                 "",
		ComputeProfileId:           profilesMap[common.PROFILE_TYPE_COMPUTE].Id,
		NetworkProfileId:           profilesMap[common.PROFILE_TYPE_NETWORK].Id,
		DatabaseParameterProfileId: profilesMap[common.PROFILE_TYPE_DATABASE_PARAMETER].Id,
	}
	// Appending request body based on database type
	appender, err := GetRequestAppender(databaseType)
	if err != nil {
		log.Error(err, "Error while getting a request appender")
		return
	}
	requestBody, err = appender.appendCloningRequest(requestBody, database, reqData)
	if err != nil {
		log.Error(err, "Error while appending clone request")
		return
	}

	log.Info("Database Cloning", "requestBody", requestBody)
	log.Info("Returning from ndb_api.GenerateCloningRequest", "database name", database.GetName())
	return
}

func (a *MSSQLRequestAppender) appendCloningRequest(req *DatabaseCloneRequest, database DatabaseInterface, reqData map[string]interface{}) (*DatabaseCloneRequest, error) {
	return nil, nil
}

func (a *MongoDbRequestAppender) appendCloningRequest(req *DatabaseCloneRequest, database DatabaseInterface, reqData map[string]interface{}) (*DatabaseCloneRequest, error) {
	return nil, nil
}

func (a *PostgresRequestAppender) appendCloningRequest(req *DatabaseCloneRequest, database DatabaseInterface, reqData map[string]interface{}) (*DatabaseCloneRequest, error) {
	req.SSHPublicKey = reqData[common.NDB_PARAM_SSH_PUBLIC_KEY].(string)

	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)
	actionArgs := []ActionArgument{
		{
			Name:  "vm_name",
			Value: database.GetName(),
		},
		{
			Name:  "dbserver_description",
			Value: "DB Server VM for " + database.GetName(),
		},
		{
			Name:  "db_password",
			Value: dbPassword,
		},
	}

	req.ActionArguments = append(req.ActionArguments, actionArgs...)
	return req, nil
}

func (a *MySqlRequestAppender) appendCloningRequest(req *DatabaseCloneRequest, database DatabaseInterface, reqData map[string]interface{}) (*DatabaseCloneRequest, error) {
	return nil, nil
}
