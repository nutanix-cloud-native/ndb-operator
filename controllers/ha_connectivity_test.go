/*
Copyright 2022-2023 Nutanix, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
)

// mockNDBClient satisfies ndb_client.NDBClientHTTPInterface for unit tests
// that exercise the ResolveIPs fallback path (DPC lookup).
type mockNDBClient struct {
	mock.Mock
}

func (m *mockNDBClient) NewRequest(method, endpoint string, body interface{}) (*http.Request, error) {
	args := m.Called(method, endpoint, body)
	if r := args.Get(0); r != nil {
		return r.(*http.Request), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockNDBClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if r := args.Get(0); r != nil {
		return r.(*http.Response), args.Error(1)
	}
	return nil, args.Error(1)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func haproxyNodeWithProperty(ip string) ndb_api.DatabaseNode {
	return ndb_api.DatabaseNode{
		Properties: []ndb_api.Property{{Name: "node_type", Value: common.HA_NODE_TYPE_HAPROXY}},
		DbServer:   ndb_api.DatabaseServer{IPAddresses: []string{ip}},
	}
}

func haproxyNodeByName(name, ip string) ndb_api.DatabaseNode {
	return ndb_api.DatabaseNode{
		DbServer: ndb_api.DatabaseServer{Name: name, IPAddresses: []string{ip}},
	}
}

func dbNode(ip string) ndb_api.DatabaseNode {
	return ndb_api.DatabaseNode{
		DbServer: ndb_api.DatabaseServer{IPAddresses: []string{ip}},
	}
}

// ---------------------------------------------------------------------------
// PostgresHAConnectivityManager
// ---------------------------------------------------------------------------

func TestPostgresHAConnectivityManager_PrimaryPort(t *testing.T) {
	mgr := &PostgresHAConnectivityManager{}

	t.Run("returns configured WritePort when set", func(t *testing.T) {
		haConfig := &ndbv1alpha1.InstanceHAConfig{
			Postgres: &ndbv1alpha1.PostgresHAConfig{WritePort: 5100},
		}
		assert.Equal(t, int32(5100), mgr.PrimaryPort(haConfig))
	})

	t.Run("returns default write port when Postgres is nil", func(t *testing.T) {
		haConfig := &ndbv1alpha1.InstanceHAConfig{Postgres: nil}
		assert.Equal(t, common.HA_PROXY_DEFAULT_WRITE_PORT, mgr.PrimaryPort(haConfig))
	})

	t.Run("returns default write port when WritePort is zero", func(t *testing.T) {
		haConfig := &ndbv1alpha1.InstanceHAConfig{
			Postgres: &ndbv1alpha1.PostgresHAConfig{WritePort: 0},
		}
		assert.Equal(t, common.HA_PROXY_DEFAULT_WRITE_PORT, mgr.PrimaryPort(haConfig))
	})
}

func TestPostgresHAConnectivityManager_AdditionalServices(t *testing.T) {
	mgr := &PostgresHAConnectivityManager{}

	t.Run("returns single -ro-svc with configured ReadPort", func(t *testing.T) {
		haConfig := &ndbv1alpha1.InstanceHAConfig{
			Postgres: &ndbv1alpha1.PostgresHAConfig{ReadPort: 5200},
		}
		svcs := mgr.AdditionalServices(haConfig)
		assert.Len(t, svcs, 1)
		assert.Equal(t, "-ro-svc", svcs[0].NameSuffix)
		assert.Equal(t, int32(5200), svcs[0].Port)
	})

	t.Run("returns -ro-svc with default read port when Postgres is nil", func(t *testing.T) {
		haConfig := &ndbv1alpha1.InstanceHAConfig{Postgres: nil}
		svcs := mgr.AdditionalServices(haConfig)
		assert.Len(t, svcs, 1)
		assert.Equal(t, "-ro-svc", svcs[0].NameSuffix)
		assert.Equal(t, common.HA_PROXY_DEFAULT_READ_PORT, svcs[0].Port)
	})

	t.Run("returns -ro-svc with default read port when ReadPort is zero", func(t *testing.T) {
		haConfig := &ndbv1alpha1.InstanceHAConfig{
			Postgres: &ndbv1alpha1.PostgresHAConfig{ReadPort: 0},
		}
		svcs := mgr.AdditionalServices(haConfig)
		assert.Len(t, svcs, 1)
		assert.Equal(t, common.HA_PROXY_DEFAULT_READ_PORT, svcs[0].Port)
	})
}

// ---------------------------------------------------------------------------
// PostgresHAIPResolver – collectHAProxyIPs
// ---------------------------------------------------------------------------

func TestPostgresHAIPResolver_collectHAProxyIPs(t *testing.T) {
	r := &PostgresHAIPResolver{}

	t.Run("identifies HAProxy node by node_type property", func(t *testing.T) {
		nodes := []ndb_api.DatabaseNode{
			haproxyNodeWithProperty("10.0.0.1"),
			dbNode("10.0.0.2"),
		}
		ips := r.collectHAProxyIPs(nodes)
		assert.Equal(t, []string{"10.0.0.1"}, ips)
	})

	t.Run("identifies HAProxy node by name containing 'haproxy' (case-insensitive)", func(t *testing.T) {
		nodes := []ndb_api.DatabaseNode{
			haproxyNodeByName("HAProxy-01", "10.0.0.3"),
			dbNode("10.0.0.4"),
		}
		ips := r.collectHAProxyIPs(nodes)
		assert.Equal(t, []string{"10.0.0.3"}, ips)
	})

	t.Run("collects IPs from multiple HAProxy nodes", func(t *testing.T) {
		nodes := []ndb_api.DatabaseNode{
			haproxyNodeWithProperty("10.0.0.1"),
			haproxyNodeByName("haproxy-02", "10.0.0.2"),
			dbNode("10.0.0.3"),
		}
		ips := r.collectHAProxyIPs(nodes)
		assert.ElementsMatch(t, []string{"10.0.0.1", "10.0.0.2"}, ips)
	})

	t.Run("excludes HAProxy node with no IP addresses", func(t *testing.T) {
		nodes := []ndb_api.DatabaseNode{
			{
				Properties: []ndb_api.Property{{Name: "node_type", Value: common.HA_NODE_TYPE_HAPROXY}},
				DbServer:   ndb_api.DatabaseServer{IPAddresses: []string{}},
			},
			dbNode("10.0.0.2"),
		}
		ips := r.collectHAProxyIPs(nodes)
		assert.Empty(t, ips)
	})

	t.Run("returns empty slice when no HAProxy nodes exist", func(t *testing.T) {
		nodes := []ndb_api.DatabaseNode{
			dbNode("10.0.0.1"),
			dbNode("10.0.0.2"),
		}
		ips := r.collectHAProxyIPs(nodes)
		assert.Empty(t, ips)
	})

	t.Run("returns empty slice for an empty node list", func(t *testing.T) {
		ips := r.collectHAProxyIPs([]ndb_api.DatabaseNode{})
		assert.Empty(t, ips)
	})

	t.Run("property match takes precedence; name match is not double-counted", func(t *testing.T) {
		// Node has both the property and "haproxy" in its name — must appear only once.
		nodes := []ndb_api.DatabaseNode{
			{
				Properties: []ndb_api.Property{{Name: "node_type", Value: common.HA_NODE_TYPE_HAPROXY}},
				DbServer:   ndb_api.DatabaseServer{Name: "haproxy-primary", IPAddresses: []string{"10.0.0.5"}},
			},
		}
		ips := r.collectHAProxyIPs(nodes)
		assert.Equal(t, []string{"10.0.0.5"}, ips)
	})
}

// ---------------------------------------------------------------------------
// PostgresHAIPResolver – ResolveIPs
// ---------------------------------------------------------------------------

func TestPostgresHAIPResolver_ResolveIPs(t *testing.T) {
	ctx := context.Background()
	r := &PostgresHAIPResolver{}

	t.Run("returns IPs directly from nodes when HAProxy property is present (no NDB call)", func(t *testing.T) {
		db := ndb_api.DatabaseResponse{
			DatabaseNodes: []ndb_api.DatabaseNode{
				haproxyNodeWithProperty("10.0.0.1"),
				dbNode("10.0.0.2"),
			},
		}
		// No NDB client call expected — passing nil intentionally.
		ips, err := r.ResolveIPs(ctx, nil, db)
		assert.NoError(t, err)
		assert.Equal(t, []string{"10.0.0.1"}, ips)
	})

	t.Run("returns IPs directly from nodes when HAProxy name match found (no NDB call)", func(t *testing.T) {
		db := ndb_api.DatabaseResponse{
			DatabaseNodes: []ndb_api.DatabaseNode{
				haproxyNodeByName("haproxy-01", "10.0.0.7"),
				dbNode("10.0.0.8"),
			},
		}
		ips, err := r.ResolveIPs(ctx, nil, db)
		assert.NoError(t, err)
		assert.Equal(t, []string{"10.0.0.7"}, ips)
	})

	t.Run("returns nil when no HAProxy nodes found and DatabaseServerId is empty", func(t *testing.T) {
		db := ndb_api.DatabaseResponse{
			DatabaseNodes: []ndb_api.DatabaseNode{
				{DatabaseServerId: "", DbServer: ndb_api.DatabaseServer{IPAddresses: []string{"10.0.0.3"}}},
			},
		}
		ips, err := r.ResolveIPs(ctx, nil, db)
		assert.NoError(t, err)
		assert.Nil(t, ips)
	})

	t.Run("propagates NDB client error from DPC fallback lookup", func(t *testing.T) {
		db := ndb_api.DatabaseResponse{
			DatabaseNodes: []ndb_api.DatabaseNode{
				{DatabaseServerId: "server-id-123", DbServer: ndb_api.DatabaseServer{IPAddresses: []string{"10.0.0.4"}}},
			},
		}
		ndbClient := &mockNDBClient{}
		// GetHAProxyIPsForCluster makes an HTTP request; simulate a transport-level failure.
		ndbClient.On("NewRequest", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("connection refused"))

		ips, err := r.ResolveIPs(ctx, ndbClient, db)
		assert.Error(t, err)
		assert.Nil(t, ips)
		ndbClient.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// Registry lookup
// ---------------------------------------------------------------------------

func TestHAConnectivityManagersRegistry(t *testing.T) {
	t.Run("postgres is registered", func(t *testing.T) {
		mgr, ok := haConnectivityManagers[common.DATABASE_TYPE_POSTGRES]
		assert.True(t, ok)
		assert.IsType(t, &PostgresHAConnectivityManager{}, mgr)
	})

	t.Run("unknown engine type is not registered", func(t *testing.T) {
		_, ok := haConnectivityManagers["mysql"]
		assert.False(t, ok)
	})
}

func TestHAIPResolversRegistry(t *testing.T) {
	t.Run("postgres is registered", func(t *testing.T) {
		res, ok := haIPResolvers[common.DATABASE_TYPE_POSTGRES]
		assert.True(t, ok)
		assert.IsType(t, &PostgresHAIPResolver{}, res)
	})

	t.Run("unknown engine type is not registered", func(t *testing.T) {
		_, ok := haIPResolvers["mysql"]
		assert.False(t, ok)
	})
}
