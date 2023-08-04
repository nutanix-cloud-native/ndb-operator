package automation

// Paths
const (
	DbSecretPath  = "../config/db-secret.yaml"
	NdbSecretPath = "../config/ndb-secret.yaml"
	DbPath        = "./database.yaml"
	AppPodPath    = "../config/pod.yaml"
	AppSvcPath    = "../config/service.yaml"
)

// PostgreSql Single Instance Yaml Info
const (
	PgSiDbSecretName  = "pg-si-db-secret-name"
	PgSiNdbSecretName = "pg-si-ndb-secret-name"
	PgSiDbName        = "mazin-pg-si-db-name"
	PgSiPodName       = "pg-si-best-app"
	PgSiSvcName       = PgSiDbName + "-svc"
	PgSiSvcPort       = 30000
)

// Microsoft SQL Server Single instance Yaml Info
const (
	MssqlSiDbSecretName  = "mssql-si-db-secret-name"
	MssqlSiNdbSecretName = "mssql-si-ndb-secret-name"
	MssqlSiDbName        = "mazin-mssql-si-db-name"
	MssqlSiPodName       = "mssql-si-best-app" // Will also be used as pod label and service selector
	MssqlSiSvcName       = MssqlSiDbName + "-svc"
	MssqlSvcPort         = 30001
)

// Oracle Single instance Yaml Info
const (
	OracleSiDbSecretName  = "oracle-si-db-secret-name"
	OracleSiNdbSecretName = "oracle-si-ndb-secret-name"
	OracleSiDbName        = "mazin-oracle-si-db-name"
	OracleSiPodName       = "oracle-si-best-app" // Will also be used as pod label and service selector
	OracleSiSvcName       = OracleSiDbName + "-svc"
	OracleSvcPort         = 30002
)

// MySQL Server Single instance Yaml Info
const (
	MysqlSiDbSecretName  = "mysql-si-db-secret-name"
	MysqlSiNdbSecretName = "mysql-si-ndb-secret-name"
	MysqlSiDbName        = "mazin-mysql-si-db-name"
	MysqlSiPodName       = "mysql-si-best-app" // Will also be used as pod label and service selector
	MysqlSiSvcName       = MysqlSiDbName + "-svc"
	MysqlSvcPort         = 30003
)

// MariaDB Server Single instance Yaml Info
const (
	MariaSiDbSecretName  = "mariaDb-si-db-secret-name"
	MariaSiNdbSecretName = "mariaDb-si-ndb-secret-name"
	MariaSiDbName        = "mazin-mariaDb-si-db-name"
	MariaSiPodName       = "mariaDb-si-best-app" // Will also be used as pod label and service selector
	MariaSiSvcName       = MariaSiDbName + "-svc"
	MariaSvcPort         = 30004
)

// MongoDb Server Single instance Yaml Info
const (
	MongoSiDbSecretName  = "mssql-si-db-secret-name"
	MongoSiNdbSecretName = "mssql-si-ndb-secret-name"
	MongoSiDbName        = "mazin-mssql-si-db-name"
	MongoSiPodName       = "mssql-si-best-app" // Will also be used as pod label and service selector
	MongoSiSvcName       = MongoSiDbName + "-svc"
	MongoSvcPort         = 30005
)
