package automation

const (
	NAMESPACE_DEFAULT = "default"

	NDBSERVER_PATH  = "./config/ndb.yaml"
	DATABASE_PATH   = "./config/database.yaml"
	DB_SECRET_PATH  = "./config/db-secret.yaml"
	NDB_SECRET_PATH = "./config/ndb-secret.yaml"
	APP_POD_PATH    = "./config/pod.yaml"
	APP_SVC_PATH    = "./config/service.yaml"

	KUBECONFIG_ENV          = "KUBECONFIG"
	DB_SECRET_PASSWORD_ENV  = "DB_SECRET_PASSWORD"
	NDB_SECRET_USERNAME_ENV = "NDB_SECRET_USERNAME"
	NDB_SECRET_PASSWORD_ENV = "NDB_SECRET_PASSWORD"
	NDB_SERVER_ENV          = "NDB_SERVER"
	NDB_CLUSTER_ID_ENV      = "NDB_CLUSTER_ID"

	PROVISIONING_LOG_PATH = "../../logs/provisioning"
	CLONING_LOG_PATH      = "../../logs/cloning"
)
