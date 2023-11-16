package util

import (
	"context"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetAllDataFromSecret(t *testing.T) {
	type args struct {
		ctx       context.Context
		client    client.Client
		name      string
		namespace string
	}
	fakeClient := fake.NewClientBuilder().WithObjects(&v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-namespace",
			Name:      "test-secret",
		},
		Data: map[string][]byte{
			"key1": []byte("value1"),
			"key2": []byte("value2"),
		},
	}).Build()

	tests := []struct {
		name     string
		args     args
		wantData map[string][]byte
		wantErr  bool
	}{
		{
			name: "Test 1: GetAllDataFromSecret fetches all the data from the secret",
			args: args{
				ctx:       context.TODO(),
				client:    fakeClient,
				name:      "test-secret",
				namespace: "test-namespace",
			},
			wantData: map[string][]byte{
				"key1": []byte("value1"),
				"key2": []byte("value2"),
			},
			wantErr: false,
		},
		{
			name: "Test 2: GetAllDataFromSecret fetches returns an error if the secret does not exist",
			args: args{
				ctx:       context.TODO(),
				client:    fakeClient,
				name:      "test-does-not-exist",
				namespace: "test-namespace",
			},
			wantData: nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotData, err := GetAllDataFromSecret(tt.args.ctx, tt.args.client, tt.args.name, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllDataFromSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotData, tt.wantData) {
				t.Errorf("GetAllDataFromSecret() = %v, want %v", gotData, tt.wantData)
			}
		})
	}
}
