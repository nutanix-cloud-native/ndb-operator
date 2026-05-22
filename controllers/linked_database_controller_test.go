package controllers

import (
	"testing"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
)

func TestResolveLinkedDatabaseSourceDatabaseId(t *testing.T) {
	ndbServer := &ndbv1alpha1.NDBServer{
		Status: ndbv1alpha1.NDBServerStatus{
			Databases: map[string]ndbv1alpha1.NDBServerDatabaseInfo{
				"postgres-id": {
					Id:   "postgres-id",
					Name: "existing-postgres",
				},
			},
		},
	}

	tests := []struct {
		name         string
		linked       *ndbv1alpha1.LinkedDatabase
		wantSourceId string
		wantErr      bool
	}{
		{
			name: "uses explicit source database id",
			linked: &ndbv1alpha1.LinkedDatabase{
				Spec: ndbv1alpha1.LinkedDatabaseSpec{
					SourceDatabaseId: "explicit-id",
				},
			},
			wantSourceId: "explicit-id",
		},
		{
			name: "resolves source database name from NDBServer status",
			linked: &ndbv1alpha1.LinkedDatabase{
				Spec: ndbv1alpha1.LinkedDatabaseSpec{
					SourceDatabaseName: "existing-postgres",
				},
			},
			wantSourceId: "postgres-id",
		},
		{
			name: "returns an error when no source database reference is provided",
			linked: &ndbv1alpha1.LinkedDatabase{
				Spec: ndbv1alpha1.LinkedDatabaseSpec{},
			},
			wantErr: true,
		},
		{
			name: "returns an error when the source database name is not in NDBServer status",
			linked: &ndbv1alpha1.LinkedDatabase{
				Spec: ndbv1alpha1.LinkedDatabaseSpec{
					SourceDatabaseName: "missing-postgres",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSourceId, err := resolveLinkedDatabaseSourceDatabaseId(tt.linked, ndbServer)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveLinkedDatabaseSourceDatabaseId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotSourceId != tt.wantSourceId {
				t.Errorf("resolveLinkedDatabaseSourceDatabaseId() = %v, want %v", gotSourceId, tt.wantSourceId)
			}
		})
	}
}
