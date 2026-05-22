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

func TestGenerateLinkedDatabaseProvisionRequest(t *testing.T) {
	tests := []struct {
		name         string
		databaseName string
		want         *LinkedDatabaseProvisionRequest
		wantErr      bool
	}{
		{
			name:         "returns an error for an empty database name",
			databaseName: "",
			wantErr:      true,
		},
		{
			name:         "creates the NDB linked database payload",
			databaseName: "test",
			want: &LinkedDatabaseProvisionRequest{
				Databases: []LinkedDatabaseRequest{{DatabaseName: "test"}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateLinkedDatabaseProvisionRequest(tt.databaseName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateLinkedDatabaseProvisionRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateLinkedDatabaseProvisionRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvisionLinkedDatabase(t *testing.T) {
	type args struct {
		ctx              context.Context
		ndbClient        ndb_client.NDBClientHTTPInterface
		sourceDatabaseId string
		req              *LinkedDatabaseProvisionRequest
	}

	provisionRequest := &LinkedDatabaseProvisionRequest{
		Databases: []LinkedDatabaseRequest{{DatabaseName: "test"}},
	}

	mockNDBClient := &MockNDBClientHTTPInterface{}
	mockNDBClient.On("NewRequest", http.MethodPost, "databases/databaseid/linked-databases", provisionRequest).Once().Return(nil, errors.New("mock-error-new-request"))

	req := &http.Request{}
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"name":"linked-db-create","operationId":"operation-id","entityId":"databaseid"}`)),
	}
	mockNDBClient.On("NewRequest", http.MethodPost, "databases/databaseid/linked-databases", provisionRequest).Once().Return(req, nil)
	mockNDBClient.On("Do", req).Once().Return(res, nil)

	tests := []struct {
		name     string
		args     args
		wantTask *TaskInfoSummaryResponse
		wantErr  bool
	}{
		{
			name: "returns an error when source database id is empty",
			args: args{
				ctx:              context.TODO(),
				ndbClient:        mockNDBClient,
				sourceDatabaseId: "",
				req:              provisionRequest,
			},
			wantErr: true,
		},
		{
			name: "returns an error when sendRequest returns an error",
			args: args{
				ctx:              context.TODO(),
				ndbClient:        mockNDBClient,
				sourceDatabaseId: "databaseid",
				req:              provisionRequest,
			},
			wantErr: true,
		},
		{
			name: "posts to the linked-databases endpoint",
			args: args{
				ctx:              context.TODO(),
				ndbClient:        mockNDBClient,
				sourceDatabaseId: "databaseid",
				req:              provisionRequest,
			},
			wantTask: &TaskInfoSummaryResponse{
				Name:        "linked-db-create",
				OperationId: "operation-id",
				EntityId:    "databaseid",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTask, err := ProvisionLinkedDatabase(tt.args.ctx, tt.args.ndbClient, tt.args.sourceDatabaseId, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProvisionLinkedDatabase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTask, tt.wantTask) {
				t.Errorf("ProvisionLinkedDatabase() = %v, want %v", gotTask, tt.wantTask)
			}
		})
	}
}

func TestGetLinkedDatabaseByName(t *testing.T) {
	type args struct {
		ctx              context.Context
		ndbClient        ndb_client.NDBClientHTTPInterface
		sourceDatabaseId string
		databaseName     string
	}

	mockNDBClient := &MockNDBClientHTTPInterface{}
	req := &http.Request{Method: http.MethodGet}
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(bytes.NewBufferString(
			`{"id":"databaseid","name":"postgres","linkedDatabases":[{"id":"linked-id","databaseName":"test","databaseStatus":"READY"}]}`,
		)),
	}
	mockNDBClient.On("NewRequest", http.MethodGet, "databases/databaseid?detailed=true", nil).Once().Return(req, nil)
	mockNDBClient.On("Do", req).Once().Return(res, nil)

	tests := []struct {
		name               string
		args               args
		wantLinkedDatabase *LinkedDatabaseResponse
		wantErr            bool
	}{
		{
			name: "returns an error when source database id is empty",
			args: args{
				ctx:              context.TODO(),
				ndbClient:        mockNDBClient,
				sourceDatabaseId: "",
				databaseName:     "test",
			},
			wantErr: true,
		},
		{
			name: "returns an error when database name is empty",
			args: args{
				ctx:              context.TODO(),
				ndbClient:        mockNDBClient,
				sourceDatabaseId: "databaseid",
				databaseName:     "",
			},
			wantErr: true,
		},
		{
			name: "finds linked database by logical database name",
			args: args{
				ctx:              context.TODO(),
				ndbClient:        mockNDBClient,
				sourceDatabaseId: "databaseid",
				databaseName:     "test",
			},
			wantLinkedDatabase: &LinkedDatabaseResponse{
				Id:             "linked-id",
				DatabaseName:   "test",
				DatabaseStatus: "READY",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLinkedDatabase, err := GetLinkedDatabaseByName(tt.args.ctx, tt.args.ndbClient, tt.args.sourceDatabaseId, tt.args.databaseName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLinkedDatabaseByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotLinkedDatabase, tt.wantLinkedDatabase) {
				t.Errorf("GetLinkedDatabaseByName() = %v, want %v", gotLinkedDatabase, tt.wantLinkedDatabase)
			}
		})
	}
}
