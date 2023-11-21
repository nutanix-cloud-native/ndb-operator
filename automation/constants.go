package automation

const (
	NAMESPACE_DEFAULT = "default"

	// Resource paths
	NDBSERVER_PATH  = "./config/ndb.yaml"
	DATABASE_PATH   = "./config/database.yaml"
	DB_SECRET_PATH  = "./config/db-secret.yaml"
	NDB_SECRET_PATH = "./config/ndb-secret.yaml"
	APP_POD_PATH    = "./config/pod.yaml"
	APP_SVC_PATH    = "./config/service.yaml"

	// Environment Variables
	KUBECONFIG_ENV          = "KUBECONFIG"
	DB_SECRET_PASSWORD_ENV  = "DB_SECRET_PASSWORD"
	NDB_SECRET_USERNAME_ENV = "NDB_SECRET_USERNAME"
	NDB_SECRET_PASSWORD_ENV = "NDB_SECRET_PASSWORD"
	NDB_SERVER_ENV          = "NDB_SERVER"
	NX_CLUSTER_ID_ENV       = "NX_CLUSTER_ID"

	MONGO_SI_CLONING_NAME_ENV    = "MONGO_SI_CLONING_NAME"
	MSSQL_SI_CLONING_NAME_ENV    = "MSSQL_SI_CLONING_NAME"
	MYSQL_SI_CLONING_NAME_ENV    = "MYSQL_SI_CLONING_NAME"
	POSTGRES_SI_CLONING_NAME_ENV = "POSTGRES_SI_CLONING_NAME"

	// Log paths
	PROVISIONING_LOG_PATH = "../../logs/provisioning"
	CLONING_LOG_PATH      = "../../logs/cloning"

	// Provisioning ports for app connectivity tests
	MONGO_SI_PROVISONING_LOCAL_PORT    = "3000"
	MSSQL_SI_PROVISONING_LOCAL_PORT    = "3001"
	MYSQL_SI_PROVISONING_LOCAL_PORT    = "3002"
	POSTGRES_SI_PROVISONING_LOCAL_PORT = "3003"

	// Cloning ports for app connectivity tests
	MONGO_SI_CLONING_LOCAL_PORT    = "3004"
	MSSQL_SI_CLONING_LOCAL_PORT    = "3005"
	MYSQL_SI_CLONING_LOCAL_PORT    = "3006"
	POSTGRES_SI_CLONING_LOCAL_PORT = "3007"

	// Clone source database names
	MONGO_SI_CLONING_NAME_DEFAULT    = "operator-mongo"
	MSSQL_SI_CLONING_NAME_DEFAULT    = "operator-mssql"
	MYSQL_SI_CLONING_NAME_DEFAULT    = "operator-mysql"
	POSTGRES_SI_CLONING_NAME_DEFAULT = "operator-postgres"
)
