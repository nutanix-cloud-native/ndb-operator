package ndb_api

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
)

// Tests the GetSLAByName function, tests the following cases:
// 1. SLA for a given name exists.
// 2. SLA for a given name does not exist.
// 3. Unable to fetch all SLAs to filter.
func TestGetSLAByName(t *testing.T) {
	SLA_RESPONSES := getMockSLAResponses()
	tests := []struct {
		slaName             string
		responseMap         map[string]interface{}
		expectedSLAResponse SLAResponse
		expectedError       error
	}{
		{
			slaName: "SLA 1",
			responseMap: map[string]interface{}{
				"GET /slas": SLA_RESPONSES,
			},
			expectedSLAResponse: SLA_RESPONSES[0],
			expectedError:       nil,
		},
		{
			slaName: "SLA-NOT-PRESENT",
			responseMap: map[string]interface{}{
				"GET /slas": SLA_RESPONSES,
			},
			expectedSLAResponse: SLAResponse{},
			expectedError:       fmt.Errorf("SLA SLA-NOT-PRESENT not found"),
		},
		{
			slaName: "SLA-X",
			responseMap: map[string]interface{}{
				"GET /slas": nil,
			},
			expectedSLAResponse: SLAResponse{},
			expectedError:       fmt.Errorf("SLA SLA-X not found"),
		},
	}

	for _, tc := range tests {
		server := GetServerTestHelperWithResponseMap(t, tc.responseMap)
		defer server.Close()
		ndb_client := ndb_client.NewNDBClient("username", "password", server.URL, "", true)

		sla, err := GetSLAByName(context.Background(), ndb_client, tc.slaName)
		if !reflect.DeepEqual(tc.expectedSLAResponse, sla) {
			t.Fatalf("expected: %v, got: %v", tc.expectedSLAResponse, sla)
		}
		if tc.expectedError != err && tc.expectedError.Error() != err.Error() {
			t.Fatalf("expected: %v, got: %v", tc.expectedError, err)
		}
	}

}
