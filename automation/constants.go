package automation

const (
	NAMESPACE_DEFAULT = "default"
	// NDB_CREDENTIALS_NAMESPACE is the dedicated namespace for NDB API credential secrets.
	// Keeping credentials here allows Database resources (in other namespaces) to reference
	// cluster-scoped NDBServer without needing access to the secret.
	NDB_CREDENTIALS_NAMESPACE = "ndb-credentials"

	// Environment Variables
	KUBECONFIG_ENV          = "KUBECONFIG"
	DB_SECRET_PASSWORD_ENV  = "DB_SECRET_PASSWORD"
	NDB_SECRET_USERNAME_ENV = "NDB_SECRET_USERNAME"
	NDB_SECRET_PASSWORD_ENV = "NDB_SECRET_PASSWORD"
	NDB_SERVER_ENV          = "NDB_SERVER"
	NX_CLUSTER_ID_ENV       = "NX_CLUSTER_ID"
	NX_CLUSTER_NAME_ENV     = "NX_CLUSTER_NAME"

	MONGO_SI_CLONING_NAME_ENV    = "MONGO_SI_CLONING_NAME"
	MSSQL_SI_CLONING_NAME_ENV    = "MSSQL_SI_CLONING_NAME"
	MYSQL_SI_CLONING_NAME_ENV    = "MYSQL_SI_CLONING_NAME"
	POSTGRES_SI_CLONING_NAME_ENV = "POSTGRES_SI_CLONING_NAME"
	ORACLE_SI_CLONING_NAME_ENV   = "ORACLE_SI_CLONING_NAME"

	// Log paths
	PROVISIONING_LOG_PATH = "../../logs/provisioning"
	CLONING_LOG_PATH      = "../../logs/cloning"

	// Provisioning ports for app connectivity tests
	MONGO_SI_PROVISONING_LOCAL_PORT    = "3000"
	MSSQL_SI_PROVISONING_LOCAL_PORT    = "3001"
	MYSQL_SI_PROVISONING_LOCAL_PORT    = "3002"
	POSTGRES_SI_PROVISONING_LOCAL_PORT = "3003"
	ORACLE_SI_PROVISONING_LOCAL_PORT   = "3008"

	// Cloning ports for app connectivity tests
	MONGO_SI_CLONING_LOCAL_PORT    = "3004"
	MSSQL_SI_CLONING_LOCAL_PORT    = "3005"
	MYSQL_SI_CLONING_LOCAL_PORT    = "3006"
	POSTGRES_SI_CLONING_LOCAL_PORT = "3007"
	ORACLE_SI_CLONING_LOCAL_PORT   = "3009"

	// Clone source database names
	MONGO_SI_CLONING_NAME_DEFAULT    = "operator-mongo"
	MSSQL_SI_CLONING_NAME_DEFAULT    = "operator-mssql"
	MYSQL_SI_CLONING_NAME_DEFAULT    = "operator-mysql"
	POSTGRES_SI_CLONING_NAME_DEFAULT = "operator-postgres"
	ORACLE_SI_CLONING_NAME_DEFAULT   = "operator-oracle"
)
