package controller_types

import (
	"errors"
	"strings"

	"github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
)

type Database struct {
	v1alpha1.Database
}

func (d *Database) GetDBInstanceName() string {
	return d.Spec.Instance.DatabaseInstanceName
}

func (d *Database) GetDBInstanceType() string {
	return d.Spec.Instance.Type
}

func (d *Database) GetDBInstanceDatabaseNames() string {
	return strings.Join(d.Spec.Instance.DatabaseNames, ",")
}

func (d *Database) GetDBInstanceTimeZone() string {
	return d.Spec.Instance.TimeZone
}

func (d *Database) GetDBInstanceSize() int {
	return d.Spec.Instance.Size
}

func (d *Database) GetNDBClusterId() string {
	return d.Spec.NDB.ClusterId
}

func (d *Database) GetDBInstanceActionArguments() []ndb_api.ActionArgument {
	dbTypeActionArgs, _ := GetActionArgumentsByDatabaseType(d.GetDBInstanceType())
	return dbTypeActionArgs.Get(d.Spec)
}

func (d *Database) GetProfileResolvers() ndb_api.ProfileResolvers {
	profileResolvers := make(ndb_api.ProfileResolvers)

	profileResolvers[common.PROFILE_TYPE_COMPUTE] = &Profile{
		Profile:     d.Spec.Instance.Profiles.Compute,
		ProfileType: common.PROFILE_TYPE_COMPUTE,
	}
	profileResolvers[common.PROFILE_TYPE_SOFTWARE] = &Profile{
		Profile:     d.Spec.Instance.Profiles.Software,
		ProfileType: common.PROFILE_TYPE_SOFTWARE,
	}
	profileResolvers[common.PROFILE_TYPE_NETWORK] = &Profile{
		Profile:     d.Spec.Instance.Profiles.Network,
		ProfileType: common.PROFILE_TYPE_NETWORK,
	}
	profileResolvers[common.PROFILE_TYPE_DATABASE_PARAMETER] = &Profile{
		Profile:     d.Spec.Instance.Profiles.DbParam,
		ProfileType: common.PROFILE_TYPE_DATABASE_PARAMETER,
	}

	return profileResolvers

}

// Returns action arguments based on the type of database
func GetActionArgumentsByDatabaseType(databaseType string) (ndb_api.DatabaseActionArgs, error) {
	var dbTypeActionArgs ndb_api.DatabaseActionArgs
	switch databaseType {
	case common.DATABASE_TYPE_MYSQL:
		dbTypeActionArgs = &MysqlActionArgs{}
	case common.DATABASE_TYPE_POSTGRES:
		dbTypeActionArgs = &PostgresActionArgs{}
	case common.DATABASE_TYPE_MONGODB:
		dbTypeActionArgs = &MongodbActionArgs{}
	default:
		return nil, errors.New("invalid database type: supported values: mysql, postgres, mongodb")
	}
	return dbTypeActionArgs, nil
}
