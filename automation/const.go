package automation

// PostgreSql Single Instance
const (
	PgSiDbSecretName  = "pg-si-db-secret-name"
	PgSiNdbSecretName = "pg-si-ndb-secret-name"
	PgSiDbName        = "mazin-pg-si-db-name"
	PgSiPodName       = "pg-si-best-app"
	PgSiSvcName       = PgSiDbName + "-svc"
	PgSiSvcPort       = 30000
	PgSiTag           = PgSiDbName + "-tag"
)

// Microsoft SQL Server Single instance
const (
	MssqlSiDbSecretName  = "mssql-si-db-secret-name"
	MssqlSiNdbSecretName = "mssql-si-ndb-secret-name"
	MssqlSiDbName        = "mazin-mssql-si-db-name"
	MssqlSiPodName       = "mssql-si-best-app"
	MssqlSiSvcName       = MssqlSiDbName + "-svc"
	MssqlSvcPort         = 30001
	MssqlSiTag           = MssqlSiDbName + "-tag"
)

// Oracle Single instance
const (
	OracleSiDbSecretName  = "oracle-si-db-secret-name"
	OracleSiNdbSecretName = "oracle-si-ndb-secret-name"
	OracleSiDbName        = "mazin-oracle-si-db-name"
	OracleSiPodName       = "oracle-si-best-app"
	OracleSiSvcName       = OracleSiDbName + "-svc"
	OracleSvcPort         = 30002
	OracleSiTag           = OracleSiDbName + "-tag"
)

// MySQL Server Single instance
const (
	MysqlSiDbSecretName  = "mysql-si-db-secret-name"
	MysqlSiNdbSecretName = "mysql-si-ndb-secret-name"
	MysqlSiDbName        = "mazin-mysql-si-db-name"
	MysqlSiPodName       = "mysql-si-best-app"
	MysqlSiSvcName       = MysqlSiDbName + "-svc"
	MysqlSvcPort         = 30003
	MysqlSiTag           = MysqlSiDbName + "-tag"
)

// MariaDB Server Single instance
const (
	MariaSiDbSecretName  = "mariaDb-si-db-secret-name"
	MariaSiNdbSecretName = "mariaDb-si-ndb-secret-name"
	MariaSiDbName        = "mazin-mariaDb-si-db-name"
	MariaSiPodName       = "mariaDb-si-best-app"
	MariaSiSvcName       = MariaSiDbName + "-svc"
	MariaSvcPort         = 30004
	MariaSiTag           = MariaSiDbName + "-tag"
)

// MongoDb Server Single instance
const (
	MongoSiDbSecretName  = "mssql-si-db-secret-name"
	MongoSiNdbSecretName = "mssql-si-ndb-secret-name"
	MongoSiDbName        = "mazin-mssql-si-db-name"
	MongoSiPodName       = "mssql-si-best-app"
	MongoSiSvcName       = MongoSiDbName + "-svc"
	MongoSvcPort         = 30005
	MongoSiTag           = MongoSiDbName + "-tag"
)
