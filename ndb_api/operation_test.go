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

func TestGetOperationById(t *testing.T) {
	type args struct {
		ctx       context.Context
		ndbClient ndb_client.NDBClientHTTPInterface
		id        string
	}

	// Mocks of the NDB Client interface
	mockNDBClient := &MockNDBClientHTTPInterface{}

	mockNDBClient.On("NewRequest", http.MethodGet, "operations/operationid?display=true", nil).Once().Return(nil, errors.New("mock-error-new-request"))

	req := &http.Request{Method: http.MethodGet}
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"id":"operationid", "name":"operation-1"}`)),
	}
	mockNDBClient.On("NewRequest", http.MethodGet, "operations/operationid?display=true", nil).Once().Return(req, nil)
	mockNDBClient.On("Do", req).Once().Return(res, nil)

	tests := []struct {
		name          string
		args          args
		wantOperation *OperationResponse
		wantErr       bool
	}{
		{
			name: "Test 1: GetOperationById returns an error when a request with empty id is passed to it",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
				id:        "",
			},
			wantOperation: nil,
			wantErr:       true,
		},
		{
			name: "Test 2: GetOperationById returns an error when sendRequest returns an error",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
				id:        "operationid",
			},
			wantOperation: nil,
			wantErr:       true,
		},
		{
			name: "Test 3: GetOperationById returns an OperationResponse when sendRequest returns a response without error",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
				id:        "operationid",
			},
			wantOperation: &OperationResponse{Id: "operationid", Name: "operation-1"},
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOperation, err := GetOperationById(tt.args.ctx, tt.args.ndbClient, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOperationById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotOperation, tt.wantOperation) {
				t.Errorf("GetOperationById() = %v, want %v", gotOperation, tt.wantOperation)
			}
		})
	}
}
