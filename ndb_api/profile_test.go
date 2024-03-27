package ndb_api

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
)

func TestGetAllProfiles(t *testing.T) {
	type args struct {
		ctx       context.Context
		ndbClient ndb_client.NDBClientHTTPInterface
	}

	// Mocks of the NDB Client interface
	mockNDBClient := &MockNDBClientHTTPInterface{}

	mockNDBClient.On("NewRequest", http.MethodGet, "profiles", nil).Once().Return(nil, errors.New("mock-error-new-request"))

	req := &http.Request{Method: http.MethodGet}
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(bytes.NewBufferString(
			`[{"id":"1", "name":"profile-1"},{"id":"2", "name":"profile-2"}]`,
		)),
	}
	mockNDBClient.On("NewRequest", http.MethodGet, "profiles", nil).Once().Return(req, nil)
	mockNDBClient.On("Do", req).Once().Return(res, nil)

	tests := []struct {
		name         string
		args         args
		wantProfiles []ProfileResponse
		wantErr      bool
	}{
		{
			name: "Test 1: GetAllProfiles returns an error when sendRequest returns an error",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
			},
			wantProfiles: nil,
			wantErr:      true,
		},
		{
			name: "Test 2: GetAllProfiles returns a slice of profiles when sendRequest does not return an error",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
			},
			wantProfiles: []ProfileResponse{
				{
					Id:   "1",
					Name: "profile-1",
				},
				{
					Id:   "2",
					Name: "profile-2",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProfiles, err := GetAllProfiles(tt.args.ctx, tt.args.ndbClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllProfiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotProfiles, tt.wantProfiles) {
				t.Errorf("GetAllProfiles() = %v, want %v", gotProfiles, tt.wantProfiles)
			}
		})
	}
}
