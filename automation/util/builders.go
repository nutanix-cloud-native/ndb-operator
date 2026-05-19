package util

import (
	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	automation "github.com/nutanix-cloud-native/ndb-operator/automation"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const sshPublicKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCwyAhpllp2WwrUB1aO/0/DN5nIWNXJWQ3ybhuEG4U+kHl8xFFKnPOTDQtTK8UwByoSf6wqIfTr10ESAoHySOpxHk2gyVHVmUmRZ1WFiNR5tW3Q4qbq1qKpIVy1jH9ZRoTJwzg0J33W9W8SZzhM8Nj0nwuDqp6FS8ui7q9H3tgM+9bYYxETTg52NEw7jTVQx6KaZgG+p/8armoYPKh9DGhBYGY3oCmGiOYlm/phSlj3R63qghZIsBXKxeJDEs4cLolQ+9QYoRqqusdEGVCp7Ba/GtUPdBPYdTy+xuXGiALEpsCrqyUstxypHZVJEQfmqS8uy9UB8KFg2YepwhPgX1oN noname"

const defaultTMSLA = "DEFAULT_OOB_GOLD_SLA"
const defaultTMDailySnapshotTime = "12:34:56"
const defaultTMSnapshotsPerDay = 4
const defaultTMLogCatchUpFrequency = 90
const defaultTMWeeklySnapshotDay = "WEDNESDAY"
const defaultTMMonthlySnapshotDay = 24
const defaultTMQuarterlySnapshotMonth = "Jan"

func defaultTMInfo(name string) *ndbv1alpha1.DBTimeMachineInfo {
	return &ndbv1alpha1.DBTimeMachineInfo{
		Name:                   name,
		Description:            "TM provisioned by operator",
		SLAName:                defaultTMSLA,
		DailySnapshotTime:      defaultTMDailySnapshotTime,
		SnapshotsPerDay:        defaultTMSnapshotsPerDay,
		LogCatchUpFrequency:    defaultTMLogCatchUpFrequency,
		WeeklySnapshotDay:      defaultTMWeeklySnapshotDay,
		MonthlySnapshotDay:     defaultTMMonthlySnapshotDay,
		QuarterlySnapshotMonth: defaultTMQuarterlySnapshotMonth,
	}
}

func ndbSecret(name string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{
			common.SECRET_DATA_KEY_USERNAME:       "",
			common.SECRET_DATA_KEY_PASSWORD:       "",
			common.SECRET_DATA_KEY_CA_CERTIFICATE: "",
		},
	}
}

func dbSecret(name string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{
			common.SECRET_DATA_KEY_PASSWORD:       "",
			common.SECRET_DATA_KEY_SSH_PUBLIC_KEY: sshPublicKey,
		},
	}
}

func ndbServer(name, ndbSecretName string) *ndbv1alpha1.NDBServer {
	return &ndbv1alpha1.NDBServer{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: ndbv1alpha1.NDBServerSpec{
			CredentialSecretRef: ndbv1alpha1.SecretReference{
				Name:      ndbSecretName,
				Namespace: automation.NDB_CREDENTIALS_NAMESPACE,
			},
			SkipCertificateVerification: true,
			// Server is set at runtime from NDB_SERVER_ENV
		},
	}
}

// --- Provisioning Builders ---

func NewPgProvisioningSetupTypes() *SetupTypes {
	return &SetupTypes{
		NdbServer: ndbServer("ndb-pg", "ndb-secret-pg-si"),
		Database: &ndbv1alpha1.Database{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "db-pg-si",
				Namespace: automation.NAMESPACE_DEFAULT,
			},
			Spec: ndbv1alpha1.DatabaseSpec{
				NDBRef: "ndb-pg",
				Instance: &ndbv1alpha1.Instance{
					Name:                "db-pg-si",
					Type:                common.DATABASE_TYPE_POSTGRES,
					DatabaseNames:       []string{"database_one", "database_two", "database_three"},
					CredentialSecret:    "db-secret-pg-si",
					Size:                10,
					TimeZone:            common.TIMEZONE_UTC,
					TMInfo:              defaultTMInfo("db-pg-si_TM"),
					AdditionalArguments: map[string]string{},
				},
			},
		},
		NdbSecret: ndbSecret("ndb-secret-pg-si"),
		DbSecret:  dbSecret("db-secret-pg-si"),
		AppPod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "app-pg-si",
				Labels: map[string]string{"app": "app-pg-si"},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "best-app",
						Image: "manavrajvanshinx/best-app:latest",
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("512Mi"),
								corev1.ResourceCPU:    resource.MustParse("1"),
							},
						},
						Env: []corev1.EnvVar{
							{Name: "DBHOST", Value: "db-pg-si-svc"},
							{Name: "DBPORT", Value: "80"},
							{
								Name: "PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{Name: "db-secret-pg-si"},
										Key:                  common.SECRET_DATA_KEY_PASSWORD,
									},
								},
							},
						},
						Ports: []corev1.ContainerPort{{ContainerPort: 3000}},
					},
				},
				InitContainers: []corev1.Container{
					{
						Name:    "init-db",
						Image:   "busybox:1.28",
						Command: []string{"sh", "-c", "until nslookup db-pg-si-svc.$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local; do echo waiting for database service; sleep 2; done"},
					},
				},
			},
		},
	}
}

func NewMySQLProvisioningSetupTypes() *SetupTypes {
	return &SetupTypes{
		NdbServer: ndbServer("ndb-mysql", "ndb-secret-mysql-si"),
		Database: &ndbv1alpha1.Database{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "db-mysql-si",
				Namespace: automation.NAMESPACE_DEFAULT,
			},
			Spec: ndbv1alpha1.DatabaseSpec{
				NDBRef: "ndb-mysql",
				Instance: &ndbv1alpha1.Instance{
					Name:                "db-mysql-si",
					Type:                common.DATABASE_TYPE_MYSQL,
					DatabaseNames:       []string{"database_one"},
					CredentialSecret:    "db-secret-mysql-si",
					Size:                10,
					TimeZone:            common.TIMEZONE_UTC,
					TMInfo:              defaultTMInfo("db-mysql-si_TM"),
					AdditionalArguments: map[string]string{},
				},
			},
		},
		NdbSecret: ndbSecret("ndb-secret-mysql-si"),
		DbSecret:  dbSecret("db-secret-mysql-si"),
		AppPod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "app-mysql-si",
				Labels: map[string]string{"app": "app-mysql-si"},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "app-mysql-container",
						Image: "mazins/ndb-operator-mysql:latest",
						Env: []corev1.EnvVar{
							{Name: "DBHOST", Value: "db-mysql-si-svc"},
							{Name: "DATABASE", Value: "database_one"},
							{Name: "DBPORT", Value: "80"},
							{Name: "USERNAME", Value: "root"},
							{
								Name: "PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{Name: "db-secret-mysql-si"},
										Key:                  common.SECRET_DATA_KEY_PASSWORD,
									},
								},
							},
						},
						Ports: []corev1.ContainerPort{{ContainerPort: 3000}},
					},
				},
				InitContainers: []corev1.Container{
					{
						Name:    "init-db",
						Image:   "busybox:1.28",
						Command: []string{"sh", "-c", "until nslookup $(DB_HOST).$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local; do echo waiting for $(DB_HOST); sleep 2; done"},
						Env:     []corev1.EnvVar{{Name: "DB_HOST", Value: "db-mysql-si-svc"}},
					},
				},
			},
		},
	}
}

func NewMongoProvisioningSetupTypes() *SetupTypes {
	return &SetupTypes{
		NdbServer: ndbServer("ndb-mongo", "ndb-secret-mongo-si"),
		Database: &ndbv1alpha1.Database{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "db-mongo-si",
				Namespace: automation.NAMESPACE_DEFAULT,
			},
			Spec: ndbv1alpha1.DatabaseSpec{
				NDBRef: "ndb-mongo",
				Instance: &ndbv1alpha1.Instance{
					Name:                "db-mongo-si",
					Type:                common.DATABASE_TYPE_MONGODB,
					DatabaseNames:       []string{"database_one"},
					CredentialSecret:    "db-secret-mongo-si",
					Size:                10,
					TimeZone:            common.TIMEZONE_UTC,
					TMInfo:              defaultTMInfo("db-mongo-si_TM"),
					AdditionalArguments: map[string]string{},
				},
			},
		},
		NdbSecret: ndbSecret("ndb-secret-mongo-si"),
		DbSecret:  dbSecret("db-secret-mongo-si"),
		AppPod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "app-mongo-si",
				Labels: map[string]string{"app": "app-mongo-si"},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "app-mongo-container",
						Image: "mazins/ndb-operator-mongodb:latest",
						Env: []corev1.EnvVar{
							{Name: "DBHOST", Value: "db-mongo-si-svc"},
							{Name: "DBPORT", Value: "80"},
							{Name: "DATABASE", Value: "database_one"},
							{Name: "USERNAME", Value: "admin"},
							{
								Name: "PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{Name: "db-secret-mongo-si"},
										Key:                  common.SECRET_DATA_KEY_PASSWORD,
									},
								},
							},
						},
						Ports: []corev1.ContainerPort{{ContainerPort: 3000}},
					},
				},
				InitContainers: []corev1.Container{
					{
						Name:    "init-db",
						Image:   "busybox:1.28",
						Command: []string{"sh", "-c", "until nslookup $(DB_HOST).$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local; do echo waiting for $(DB_HOST); sleep 2; done"},
						Env:     []corev1.EnvVar{{Name: "DB_HOST", Value: "db-mongo-si-svc"}},
					},
				},
			},
		},
	}
}

func NewMSSQLProvisioningSetupTypes() *SetupTypes {
	return &SetupTypes{
		NdbServer: ndbServer("ndb-mssql", "ndb-secret-mssql-si"),
		Database: &ndbv1alpha1.Database{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "db-mssql-si",
				Namespace: automation.NAMESPACE_DEFAULT,
			},
			Spec: ndbv1alpha1.DatabaseSpec{
				NDBRef: "ndb-mssql",
				Instance: &ndbv1alpha1.Instance{
					Name:             "db-mssql-si",
					Type:             common.DATABASE_TYPE_MSSQL,
					DatabaseNames:    []string{"database_one"},
					CredentialSecret: "db-secret-mssql-si",
					Size:             10,
					TimeZone:         common.TIMEZONE_UTC,
					Profiles: &ndbv1alpha1.Profiles{
						DbParam:         ndbv1alpha1.Profile{Name: "DEFAULT_SQLSERVER_DATABASE_PARAMS"},
						DbParamInstance: ndbv1alpha1.Profile{Name: "DEFAULT_SQLSERVER_INSTANCE_PARAMS"},
					},
					TMInfo: defaultTMInfo("db-mssql-si_TM"),
					AdditionalArguments: map[string]string{
						"authentication_mode": "mixed",
						"sql_user_password":   "Nutanix.1",
					},
				},
			},
		},
		NdbSecret: ndbSecret("ndb-secret-mssql-si"),
		DbSecret:  dbSecret("db-secret-mssql-si"),
		AppPod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "app-mssql-si",
				Labels: map[string]string{"app": "app-mssql-si"},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "app-mssql-container",
						Image: "mazins/ndb-operator-mssql:latest",
						Env: []corev1.EnvVar{
							{Name: "DBHOST", Value: "db-mssql-si-svc"},
							{Name: "USERNAME", Value: "sa"},
							{Name: "DATABASE", Value: "database_one"},
							{Name: "DBPORT", Value: "80"},
							{Name: "MSSQL_INSTANCE_NAME", Value: "CDMINSTANCE"},
							{
								Name: "PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{Name: "db-secret-mssql-si"},
										Key:                  common.SECRET_DATA_KEY_PASSWORD,
									},
								},
							},
						},
						Ports: []corev1.ContainerPort{{ContainerPort: 3000}},
					},
				},
				InitContainers: []corev1.Container{
					{
						Name:    "init-db",
						Image:   "busybox:1.28",
						Command: []string{"sh", "-c", "until nslookup $(DB_HOST).$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local; do echo waiting for $(DB_HOST); sleep 2; done"},
						Env:     []corev1.EnvVar{{Name: "DB_HOST", Value: "db-mssql-si-svc"}},
					},
				},
			},
		},
	}
}

// --- Cloning Builders ---

func NewPgCloningSetupTypes() *SetupTypes {
	return &SetupTypes{
		NdbServer: ndbServer("clone-ndb-pg-si", "clone-ndb-secret-pg-si"),
		Database: &ndbv1alpha1.Database{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "clone-pg-si",
				Namespace: automation.NAMESPACE_DEFAULT,
			},
			Spec: ndbv1alpha1.DatabaseSpec{
				NDBRef:  "clone-ndb-pg-si",
				IsClone: true,
				Clone: &ndbv1alpha1.Clone{
					Name:             "clone-pg-si",
					Type:             common.DATABASE_TYPE_POSTGRES,
					Description:      "Cloning pg single instance testing",
					CredentialSecret: "clone-db-secret-pg-si",
					TimeZone:         common.TIMEZONE_UTC,
					// ClusterName, SourceDatabaseName set at runtime from env vars
					AdditionalArguments: map[string]string{},
				},
			},
		},
		NdbSecret: ndbSecret("clone-ndb-secret-pg-si"),
		DbSecret:  dbSecret("clone-db-secret-pg-si"),
		AppPod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "clone-app-pg-si",
				Labels: map[string]string{"app": "clone-app-pg-si"},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "best-app",
						Image: "manavrajvanshinx/best-app:latest",
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("512Mi"),
								corev1.ResourceCPU:    resource.MustParse("1"),
							},
						},
						Env: []corev1.EnvVar{
							{Name: "DBHOST", Value: "clone-pg-si-svc"},
							{Name: "DBPORT", Value: "80"},
							{
								Name: "PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{Name: "clone-db-secret-pg-si"},
										Key:                  common.SECRET_DATA_KEY_PASSWORD,
									},
								},
							},
						},
						Ports: []corev1.ContainerPort{{ContainerPort: 3000}},
					},
				},
				InitContainers: []corev1.Container{
					{
						Name:    "init-db",
						Image:   "busybox:1.28",
						Command: []string{"sh", "-c", "until nslookup clone-pg-si-svc.$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local; do echo waiting for database service; sleep 2; done"},
					},
				},
			},
		},
	}
}

func NewMySQLCloningSetupTypes() *SetupTypes {
	return &SetupTypes{
		NdbServer: ndbServer("clone-ndb-mysql-si", "clone-ndb-secret-mysql-si"),
		Database: &ndbv1alpha1.Database{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "clone-mysql-si",
				Namespace: automation.NAMESPACE_DEFAULT,
			},
			Spec: ndbv1alpha1.DatabaseSpec{
				NDBRef:  "clone-ndb-mysql-si",
				IsClone: true,
				Clone: &ndbv1alpha1.Clone{
					Name:             "clone-mysql-si",
					Type:             common.DATABASE_TYPE_MYSQL,
					Description:      "Cloning mysql single instance testing",
					CredentialSecret: "clone-db-secret-mysql-si",
					TimeZone:         common.TIMEZONE_UTC,
					// ClusterName, SourceDatabaseName set at runtime from env vars
					AdditionalArguments: map[string]string{},
				},
			},
		},
		NdbSecret: ndbSecret("clone-ndb-secret-mysql-si"),
		DbSecret:  dbSecret("clone-db-secret-mysql-si"),
		AppPod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "clone-app-mysql-si",
				Labels: map[string]string{"app": "clone-app-mysql-si"},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "app-mysql-container",
						Image: "mazins/ndb-operator-mysql:latest",
						Env: []corev1.EnvVar{
							{Name: "DBHOST", Value: "clone-mysql-si-svc"},
							{Name: "DATABASE", Value: "database_one"},
							{Name: "DBPORT", Value: "3306"},
							{Name: "USERNAME", Value: "root"},
							{
								Name: "PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{Name: "clone-db-secret-mysql-si"},
										Key:                  common.SECRET_DATA_KEY_PASSWORD,
									},
								},
							},
						},
						Ports: []corev1.ContainerPort{{ContainerPort: 3000}},
					},
				},
				InitContainers: []corev1.Container{
					{
						Name:    "init-db",
						Image:   "busybox:1.28",
						Command: []string{"sh", "-c", "until nslookup $(DB_HOST).$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local; do echo waiting for $(DB_HOST); sleep 2; done"},
						Env:     []corev1.EnvVar{{Name: "DB_HOST", Value: "clone-mysql-si-svc"}},
					},
				},
			},
		},
	}
}

func NewMongoCloningSetupTypes() *SetupTypes {
	return &SetupTypes{
		NdbServer: ndbServer("clone-ndb-mongo-si", "clone-ndb-secret-pg-si"),
		Database: &ndbv1alpha1.Database{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "clone-mongo-si",
				Namespace: automation.NAMESPACE_DEFAULT,
			},
			Spec: ndbv1alpha1.DatabaseSpec{
				NDBRef:  "clone-ndb-mongo-si",
				IsClone: true,
				Clone: &ndbv1alpha1.Clone{
					Name:             "clone-mongo-si",
					Type:             common.DATABASE_TYPE_MONGODB,
					Description:      "Cloning mongoDB single instance testing",
					CredentialSecret: "clone-db-secret-mongo-si",
					TimeZone:         "America/Los_Angeles",
					// ClusterName, SourceDatabaseName set at runtime from env vars
					AdditionalArguments: map[string]string{},
				},
			},
		},
		NdbSecret: ndbSecret("clone-ndb-secret-pg-si"),
		DbSecret:  dbSecret("clone-db-secret-mongo-si"),
		AppPod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "clone-app-mongo-si",
				Labels: map[string]string{"app": "clone-app-mongo-si"},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "app-mongo-container",
						Image: "mazins/ndb-operator-mongodb:latest",
						Env: []corev1.EnvVar{
							{Name: "DBHOST", Value: "clone-mongo-si-svc"},
							{Name: "DBPORT", Value: "80"},
							{Name: "DATABASE", Value: "database_one"},
							{Name: "USERNAME", Value: "admin"},
							{
								Name: "PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{Name: "clone-db-secret-mongo-si"},
										Key:                  common.SECRET_DATA_KEY_PASSWORD,
									},
								},
							},
						},
						Ports: []corev1.ContainerPort{{ContainerPort: 3000}},
					},
				},
				InitContainers: []corev1.Container{
					{
						Name:    "init-db",
						Image:   "busybox:1.28",
						Command: []string{"sh", "-c", "until nslookup $(DB_HOST).$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local; do echo waiting for $(DB_HOST); sleep 2; done"},
						Env:     []corev1.EnvVar{{Name: "DB_HOST", Value: "clone-mongo-si-svc"}},
					},
				},
			},
		},
	}
}

func NewMSSQLCloningSetupTypes() *SetupTypes {
	return &SetupTypes{
		NdbServer: ndbServer("clone-ndb-mssql-si", "clone-ndb-secret-mssql-si"),
		Database: &ndbv1alpha1.Database{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "clone-mssql-si",
				Namespace: automation.NAMESPACE_DEFAULT,
			},
			Spec: ndbv1alpha1.DatabaseSpec{
				NDBRef:  "clone-ndb-mssql-si",
				IsClone: true,
				Clone: &ndbv1alpha1.Clone{
					Name:        "clone-mssql-si",
					Type:        common.DATABASE_TYPE_MSSQL,
					Description: "Cloning mssql single instance testing",
					Profiles: &ndbv1alpha1.Profiles{
						Software: ndbv1alpha1.Profile{Name: "MSSQL-MAZIN2"},
					},
					CredentialSecret: "clone-db-secret-mssql-si",
					TimeZone:         common.TIMEZONE_UTC,
					// ClusterName, SourceDatabaseName set at runtime from env vars
					AdditionalArguments: map[string]string{
						"authentication_mode": "mixed",
					},
				},
			},
		},
		NdbSecret: ndbSecret("clone-ndb-secret-mssql-si"),
		DbSecret:  dbSecret("clone-db-secret-mssql-si"),
		AppPod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "clone-app-mssql-si",
				Labels: map[string]string{"app": "clone-app-mssql-si"},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "app-mssql-container",
						Image: "mazins/ndb-operator-mssql:latest",
						Env: []corev1.EnvVar{
							{Name: "DBHOST", Value: "clone-mssql-si-svc"},
							{Name: "USERNAME", Value: "sa"},
							{Name: "DATABASE", Value: "database_one"},
							{Name: "DBPORT", Value: "80"},
							{Name: "MSSQL_INSTANCE_NAME", Value: "CDMINSTANCE"},
							{
								Name: "PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{Name: "clone-db-secret-mssql-si"},
										Key:                  common.SECRET_DATA_KEY_PASSWORD,
									},
								},
							},
						},
						Ports: []corev1.ContainerPort{{ContainerPort: 3000}},
					},
				},
				InitContainers: []corev1.Container{
					{
						Name:    "init-db",
						Image:   "busybox:1.28",
						Command: []string{"sh", "-c", "until nslookup $(DB_HOST).$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local; do echo waiting for $(DB_HOST); sleep 2; done"},
						Env:     []corev1.EnvVar{{Name: "DB_HOST", Value: "clone-mssql-si-svc"}},
					},
				},
			},
		},
	}
}
