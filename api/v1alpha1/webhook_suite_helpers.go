package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/* Creates a database CR with an ndb spec field missing */
func ndbSpecMissing() *Database {
	return &Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db",
			Namespace: "default",
		},
		Spec: DatabaseSpec{
			NDB: NDB{},
			Instance: Instance{
				CredentialSecret:     "db-instance-secret",
				DatabaseInstanceName: "db-instance-name",
				Type:                 "postgres",
				Size:                 10,
				TimeZone:             "UTC",
			},
		},
	}
}

/* Creates a database CR with an ndb cluster id missing */
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
				CredentialSecret:     "db-instance-secret",
				DatabaseInstanceName: "db-instance-name",
				Type:                 "postgres",
				Size:                 10,
				TimeZone:             "UTC",
			},
		},
	}
}

/* Creates a database CR with an ndb credential secret missing */
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
				CredentialSecret:     "db-instance-secret",
				DatabaseInstanceName: "db-instance-name",
				Type:                 "postgres",
				Size:                 10,
				TimeZone:             "UTC",
			},
		},
	}
}

/* Creates a database CR with an ndb server URL missing */
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
				CredentialSecret:     "db-instance-secret",
				DatabaseInstanceName: "db-instance-name",
				Type:                 "postgres",
				Size:                 10,
				TimeZone:             "UTC",
			},
		},
	}
}
