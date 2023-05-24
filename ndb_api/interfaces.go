package ndb_api

import (
	"context"

	"github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
)

type ProfileResolver interface {
	Resolve(ctx context.Context, allProfiles []ProfileResponse, filter func(p ProfileResponse) bool) (profile ProfileResponse, err error)
}

type DatabaseActionArgs interface {
	Get(dbSpec v1alpha1.DatabaseSpec) []ActionArgument
}
