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

func TestGetAllSLAs(t *testing.T) {
	type args struct {
		ctx       context.Context
		ndbClient ndb_client.NDBClientHTTPInterface
	}

	// Mocks of the NDB Client interface
	mockNDBClient := &MockNDBClientHTTPInterface{}

	mockNDBClient.On("NewRequest", http.MethodGet, "slas", nil).Once().Return(nil, errors.New("mock-error-new-request"))

	req := &http.Request{Method: http.MethodGet}
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(bytes.NewBufferString(
			`[{"id":"1", "name":"sla-1"},{"id":"2", "name":"sla-2"}]`,
		)),
	}
	mockNDBClient.On("NewRequest", http.MethodGet, "slas", nil).Once().Return(req, nil)
	mockNDBClient.On("Do", req).Once().Return(res, nil)

	tests := []struct {
		name     string
		args     args
		wantSLAs []SLAResponse
		wantErr  bool
	}{
		{
			name: "Test 1: GetAllSLAs returns an error when sendRequest returns an error",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
			},
			wantSLAs: nil,
			wantErr:  true,
		},
		{
			name: "Test 2: GetAllSLAs returns a slice of SLAs when sendRequest does not return an error",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
			},
			wantSLAs: []SLAResponse{
				{
					Id:   "1",
					Name: "sla-1",
				},
				{
					Id:   "2",
					Name: "sla-2",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSLAs, err := GetAllSLAs(tt.args.ctx, tt.args.ndbClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllSLAs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotSLAs, tt.wantSLAs) {
				t.Errorf("GetAllSLAs() = %v, want %v", gotSLAs, tt.wantSLAs)
			}
		})
	}
}
