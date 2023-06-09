package ndb_api

import (
	"errors"

	"github.com/nutanix-cloud-native/ndb-operator/common"
)

type DBProvisionRequestAppender interface {
	appendRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) *DatabaseProvisionRequest
}

type MSSqlProvisionRequestAppender struct{}

type MongoDbProvisionRequestAppender struct{}

type PostgresProvisionRequestAppender struct{}

type MySqlProvisionRequestAppender struct{}

func (a *MSSqlProvisionRequestAppender) appendRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) *DatabaseProvisionRequest {
	req.DatabaseName = string(database.GetDBInstanceDatabaseNames())
	adminPassword := reqData[common.NDB_PARAM_PASSWORD].(string)

	actionArgs := []ActionArgument{
		{
			Name:  "working_dir",
			Value: "C:\\temp",
		},
		{
			Name:  "sql_user_name",
			Value: "sa",
		},
		{
			Name:  "authentication_mode",
			Value: "windows",
		},
		{
			Name:  "delete_vm_on_failure",
			Value: "false",
		},
		{
			Name:  "is_gmsa_sql_service_account",
			Value: "false",
		},
		{
			Name:  "provision_from_backup",
			Value: "false",
		},
		{
			Name:  "distribute_database_data",
			Value: "true",
		},
		{
			Name:  "retain_database_in_restoring_mode",
			Value: "false",
		},
		// fetch this from main
		{
			Name:  "dbserver_name",
			Value: "sql-db-server",
		},
		{
			Name:  "server_collation",
			Value: "SQL_Latin1_General_CP1_CI_AS",
		},
		{
			Name:  "database_collation",
			Value: "SQL_Latin1_General_CP1_CI_AS",
		},
		{
			Name:  "dbParameterProfileIdInstance",
			Value: "af9729f6-ef72-43ac-9e02-633ad13c8d51",
		},
		// No db_password
		{
			Name:  "vm_dbserver_admin_password",
			Value: adminPassword,
		},
	}

	req.ActionArguments = append(req.ActionArguments, actionArgs...)

	return req
}

func (a *MongoDbProvisionRequestAppender) appendRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) *DatabaseProvisionRequest {
	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)
	databaseNames := database.GetDBInstanceDatabaseNames()
	SSHPublicKey := reqData[common.NDB_PARAM_SSH_PUBLIC_KEY].(string)

	req.SSHPublicKey = SSHPublicKey
	actionArgs := []ActionArgument{
		{
			Name:  "listener_port",
			Value: "27017",
		},
		{
			Name:  "log_size",
			Value: "100",
		},
		{
			Name:  "journal_size",
			Value: "100",
		},
		{
			Name:  "restart_mongod",
			Value: "true",
		},
		{
			Name:  "working_dir",
			Value: "/tmp",
		},
		{
			Name:  "db_user",
			Value: "admin",
		},
		{
			Name:  "backup_policy",
			Value: "primary_only",
		},
		{
			Name:  "db_password",
			Value: dbPassword,
		},
		{
			Name:  "database_names",
			Value: databaseNames,
		},
	}

	req.ActionArguments = append(req.ActionArguments, actionArgs...)
	return req
}

func (a *PostgresProvisionRequestAppender) appendRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) *DatabaseProvisionRequest {
	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)
	databaseNames := database.GetDBInstanceDatabaseNames()
	SSHPublicKey := reqData[common.NDB_PARAM_SSH_PUBLIC_KEY].(string)

	req.SSHPublicKey = SSHPublicKey
	actionArgs := []ActionArgument{
		{
			Name:  "proxy_read_port",
			Value: "5001",
		},
		{
			Name:  "listener_port",
			Value: "5432",
		},
		{
			Name:  "proxy_write_port",
			Value: "5000",
		},
		{
			Name:  "enable_synchronous_mode",
			Value: "false",
		},
		{
			Name:  "auto_tune_staging_drive",
			Value: "true",
		},
		{
			Name:  "backup_policy",
			Value: "primary_only",
		},
		{
			Name:  "db_password",
			Value: dbPassword,
		},
		{
			Name:  "database_names",
			Value: databaseNames,
		},
	}

	req.ActionArguments = append(req.ActionArguments, actionArgs...)
	return req
}

func (a *MySqlProvisionRequestAppender) appendRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) *DatabaseProvisionRequest {
	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)
	databaseNames := database.GetDBInstanceDatabaseNames()
	SSHPublicKey := reqData[common.NDB_PARAM_SSH_PUBLIC_KEY].(string)

	req.SSHPublicKey = SSHPublicKey
	actionArgs := []ActionArgument{
		{
			Name:  "listener_port",
			Value: "3306",
		},
		{
			Name:  "db_password",
			Value: dbPassword,
		},
		{
			Name:  "database_names",
			Value: databaseNames,
		},
	}

	req.ActionArguments = append(req.ActionArguments, actionArgs...)
	return req
}

func GetDbProvRequestAppender(databaseType string) (requestAppender DBProvisionRequestAppender, err error) {
	//var dbProvRequestAppender DBProvisionRequestAppender
	switch databaseType {
	case common.DATABASE_TYPE_MYSQL:
		requestAppender = &MySqlProvisionRequestAppender{}
	case common.DATABASE_TYPE_POSTGRES:
		requestAppender = &PostgresProvisionRequestAppender{}
	case common.DATABASE_TYPE_MONGODB:
		requestAppender = &MongoDbProvisionRequestAppender{}
	case common.DATABASE_TYPE_MSSQL:
		requestAppender = &MSSqlProvisionRequestAppender{}
	default:
		return nil, errors.New("invalid database type: supported values: mssql, mysql, postgres, mongodb")
	}
	return
}
