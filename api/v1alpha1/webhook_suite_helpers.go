package v1alpha1

import (
	ndb_api "github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/* Creates a database CR with a ndb spec field missing. */
func ndbSpecMissing() *Database {
	return &Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db",
			Namespace: "default",
		},
		Spec: DatabaseSpec{
			NDB: NDB{},
			Instance: Instance{
				DatabaseInstanceName: "db-instance-name",
				CredentialSecret:     "db-instance-secret",
				Size:                 10,
				Type:                 "postgres",
				TimeZone:             "UTC",
			},
		},
	}
}

/* Creates a database CR with a ndb 'clusterId' missing. */
func ndbClusterIdMissing() *Database {
	return &Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db",
			Namespace: "default",
		},
		Spec: DatabaseSpec{
			NDB: NDB{
				SkipCertificateVerification: true,
				CredentialSecret:            "ndb-secret",
				Server:                      "https://10.51.140.43:8443/era/v0.9",
			},
			Instance: Instance{
				DatabaseInstanceName: "db-instance-name",
				CredentialSecret:     "db-instance-secret",
				Size:                 10,
				Type:                 "postgres",
				TimeZone:             "UTC",
			},
		},
	}
}

/* Creates a database CR with a ndb 'credentialSecret' missing. */
func ndbCredentialSecretMissing() *Database {
	return &Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db",
			Namespace: "default",
		},
		Spec: DatabaseSpec{
			NDB: NDB{
				ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
				SkipCertificateVerification: true,
				Server:                      "https://10.51.140.43:8443/era/v0.9",
			},
			Instance: Instance{
				DatabaseInstanceName: "db-instance-name",
				CredentialSecret:     "db-instance-secret",
				Size:                 10,
				Type:                 "postgres",
				TimeZone:             "UTC",
			},
		},
	}
}

/* Creates a database CR with an ndb 'server' URL missing. */
func ndbServerURLMissing() *Database {
	return &Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db",
			Namespace: "default",
		},
		Spec: DatabaseSpec{
			NDB: NDB{
				ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
				SkipCertificateVerification: true,
				CredentialSecret:            "ndb-secret",
			},
			Instance: Instance{
				DatabaseInstanceName: "db-instance-name",
				CredentialSecret:     "db-instance-secret",
				Size:                 10,
				Type:                 "postgres",
				TimeZone:             "UTC",
			},
		},
	}
}

/* Creates a database CR with a 'databaseInstanceName' missing. */
func dbInstanceNameMissing() *Database {
	return &Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db",
			Namespace: "default",
		},
		Spec: DatabaseSpec{
			NDB: NDB{
				ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
				SkipCertificateVerification: true,
				CredentialSecret:            "ndb-secret",
				Server:                      "https://10.51.140.43:8443/era/v0.9",
			},
			Instance: Instance{
				CredentialSecret: "db-instance-secret",
				Size:             10,
				TimeZone:         "UTC",
				Type:             "postgres",
			},
		},
	}
}

/* Creates a database CR with 'description' not specified. */
func dbDescriptionNotSpecified(db string) *Database {
	return &Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      db,
			Namespace: "default",
		},
		Spec: DatabaseSpec{
			NDB: NDB{
				ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
				SkipCertificateVerification: true,
				CredentialSecret:            "ndb-secret",
				Server:                      "https://10.51.140.43:8443/era/v0.9",
			},
			Instance: Instance{
				DatabaseInstanceName: "db-instance-name",
				CredentialSecret:     "db-instance-secret",
				Size:                 10,
				TimeZone:             "UTC",
				Type:                 "postgres",
			},
		},
	}
}

/* Creates a database CR with 'database_names' not specified. */
func dbDatabaseNamesNotSpecified(db string) *Database {
	return &Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      db,
			Namespace: "default",
		},
		Spec: DatabaseSpec{
			NDB: NDB{
				ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
				SkipCertificateVerification: true,
				CredentialSecret:            "ndb-secret",
				Server:                      "https://10.51.140.43:8443/era/v0.9",
			},
			Instance: Instance{
				DatabaseInstanceName: "db-instance-name",
				CredentialSecret:     "db-instance-secret",
				Size:                 10,
				TimeZone:             "UTC",
				Type:                 "postgres",
			},
		},
	}
}

/* Creates a database CR with 'credentialSecret' missing. */
func dbCredentialSecretMissing() *Database {
	return &Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db",
			Namespace: "default",
		},
		Spec: DatabaseSpec{
			NDB: NDB{
				ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
				SkipCertificateVerification: true,
				CredentialSecret:            "ndb-secret",
				Server:                      "https://10.51.140.43:8443/era/v0.9",
			},
			Instance: Instance{
				DatabaseInstanceName: "db-instance-name",
				Size:                 10,
				TimeZone:             "UTC",
				Type:                 "postgres",
			},
		},
	}
}

/* Creates a database CR with a size less than 10.*/
func dbSizeLessThan10() *Database {
	return &Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db",
			Namespace: "default",
		},
		Spec: DatabaseSpec{
			NDB: NDB{
				ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
				SkipCertificateVerification: true,
				CredentialSecret:            "ndb-secret",
				Server:                      "https://10.51.140.43:8443/era/v0.9",
			},
			Instance: Instance{
				DatabaseInstanceName: "db-instance-name",
				CredentialSecret:     "db-instance-secret",
				Size:                 1,
				TimeZone:             "UTC",
				Type:                 "postgres",
			},
		},
	}
}

/* Creates a database CR with 'timeZone' not specified. */
func dbTimeZoneNotSpecified() *Database {
	return &Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db",
			Namespace: "default",
		},
		Spec: DatabaseSpec{
			NDB: NDB{
				ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
				SkipCertificateVerification: true,
				CredentialSecret:            "ndb-secret",
				Server:                      "https://10.51.140.43:8443/era/v0.9",
			},
			Instance: Instance{
				DatabaseInstanceName: "db-instance-name",
				CredentialSecret:     "db-instance-secret",
				Size:                 1,
				Type:                 "postgres",
			},
		},
	}
}

/* Creates a database CR with 'type' missing. */
func dbTypeMissing() *Database {
	return &Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db",
			Namespace: "default",
		},
		Spec: DatabaseSpec{
			NDB: NDB{
				ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
				SkipCertificateVerification: true,
				CredentialSecret:            "ndb-secret",
				Server:                      "https://10.51.140.43:8443/era/v0.9",
			},
			Instance: Instance{
				DatabaseInstanceName: "db-instance-name",
				CredentialSecret:     "db-instance-secret",
				Size:                 1,
			},
		},
	}
}

/* Creates a database CR with 'type'. */
func dbWithType(db string, typ string) *Database {
	return &Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      db,
			Namespace: "default",
		},
		Spec: DatabaseSpec{
			NDB: NDB{
				ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
				SkipCertificateVerification: true,
				CredentialSecret:            "ndb-secret",
				Server:                      "https://10.51.140.43:8443/era/v0.9",
			},
			Instance: Instance{
				DatabaseInstanceName: "db-instance-name",
				CredentialSecret:     "db-instance-secret",
				Size:                 10,
				Type:                 typ,
			},
		},
	}
}

/* Creates a database with 'timeMachine' not specified. */
func dbTimeMachineNotSpecified(db string) *Database {
	return &Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      db,
			Namespace: "default",
		},
		Spec: DatabaseSpec{
			NDB: NDB{
				ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
				SkipCertificateVerification: true,
				CredentialSecret:            "ndb-secret",
				Server:                      "https://10.51.140.43:8443/era/v0.9",
			},
			Instance: Instance{
				DatabaseInstanceName: "db-instance-name",
				CredentialSecret:     "db-instance-secret",
				Size:                 10,
				Type:                 "postgres",
			},
		},
	}
}

/* Creates a database CR with 'db' name, 'type', and 'typeDetails' specified */
func dbWithTypeDetailsSpecified(db string, typ string, typeDetails []ndb_api.ActionArgument) *Database {
	database := &Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      db,
			Namespace: "default",
		},
		Spec: DatabaseSpec{
			NDB: NDB{
				ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
				SkipCertificateVerification: true,
				CredentialSecret:            "ndb-secret",
				Server:                      "https://10.51.140.43:8443/era/v0.9",
			},
			Instance: Instance{
				DatabaseInstanceName: "db-instance-name",
				CredentialSecret:     "db-instance-secret",
				Size:                 10,
				Type:                 typ,
				TypeDetails:          typeDetails,
			},
		},
	}

	if typ == "mssql" {
		database.Spec.Instance.Profiles = &Profiles{
			Software: Profile{Name: "MSSQL_SOFTWARE_PROFILE"},
		}
	}

	return database
}