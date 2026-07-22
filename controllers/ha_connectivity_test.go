/*
Copyright 2022-2026 Nutanix, Inc.

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
	"io"
	"net/http"
	"strings"
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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	t.Run("postgres is registered", func(t *testing.T) {
		mgr, ok := haConnectivityManagers[common.DATABASE_TYPE_POSTGRES]
		assert.True(t, ok)
		assert.IsType(t, &PostgresHAConnectivityManager{}, mgr)
	})

	t.Run("mysql is registered", func(t *testing.T) {
		mgr, ok := haConnectivityManagers[common.DATABASE_TYPE_MYSQL]
		assert.True(t, ok)
		assert.IsType(t, &MySQLHAConnectivityManager{}, mgr)
	})

	t.Run("unknown engine type is not registered", func(t *testing.T) {
		_, ok := haConnectivityManagers["oracle"]
		assert.False(t, ok)
	})
}

func TestHAIPResolversRegistry(t *testing.T) {
	t.Parallel()

	t.Run("postgres is registered", func(t *testing.T) {
		res, ok := haIPResolvers[common.DATABASE_TYPE_POSTGRES]
		assert.True(t, ok)
		assert.IsType(t, &PostgresHAIPResolver{}, res)
	})

	t.Run("mysql is registered", func(t *testing.T) {
		res, ok := haIPResolvers[common.DATABASE_TYPE_MYSQL]
		assert.True(t, ok)
		assert.IsType(t, &MySQLHAIPResolver{}, res)
	})

	t.Run("unknown engine type is not registered", func(t *testing.T) {
		_, ok := haIPResolvers["oracle"]
		assert.False(t, ok)
	})
}

// ---------------------------------------------------------------------------
// Helpers for MySQL tests
// ---------------------------------------------------------------------------

func masterDBNode(ip string) ndb_api.DatabaseNode {
	return ndb_api.DatabaseNode{
		Properties: []ndb_api.Property{{Name: "role", Value: common.HA_NODE_ROLE_MASTER}},
		DbServer:   ndb_api.DatabaseServer{IPAddresses: []string{ip}},
	}
}

func replicaDBNode(ip string) ndb_api.DatabaseNode {
	return ndb_api.DatabaseNode{
		Properties: []ndb_api.Property{{Name: "role", Value: common.HA_NODE_ROLE_REPLICA}},
		DbServer:   ndb_api.DatabaseServer{IPAddresses: []string{ip}},
	}
}

// ---------------------------------------------------------------------------
// MySQLHAConnectivityManager — PrimaryPort
// ---------------------------------------------------------------------------

func TestMySQLHAConnectivityManager_PrimaryPort(t *testing.T) {
	t.Parallel()

	mgr := &MySQLHAConnectivityManager{}

	t.Run("returns listener port 3306 when router is not deployed and MySQL config is nil", func(t *testing.T) {
		haConfig := &ndbv1alpha1.InstanceHAConfig{MySQL: nil}
		assert.Equal(t, common.HA_MYSQL_DEFAULT_LISTENER_PORT, mgr.PrimaryPort(haConfig))
	})

	t.Run("returns listener port 3306 when DeployMySQLRouter is false", func(t *testing.T) {
		haConfig := &ndbv1alpha1.InstanceHAConfig{
			MySQL: &ndbv1alpha1.MySQLHAConfig{DeployMySQLRouter: false, RouterRWPort: 6446},
		}
		assert.Equal(t, common.HA_MYSQL_DEFAULT_LISTENER_PORT, mgr.PrimaryPort(haConfig))
	})

	t.Run("returns configured RouterRWPort when router is deployed", func(t *testing.T) {
		haConfig := &ndbv1alpha1.InstanceHAConfig{
			MySQL: &ndbv1alpha1.MySQLHAConfig{DeployMySQLRouter: true, RouterRWPort: 6500},
		}
		assert.Equal(t, int32(6500), mgr.PrimaryPort(haConfig))
	})

	t.Run("returns default RW port when router is deployed but RouterRWPort is zero", func(t *testing.T) {
		haConfig := &ndbv1alpha1.InstanceHAConfig{
			MySQL: &ndbv1alpha1.MySQLHAConfig{DeployMySQLRouter: true, RouterRWPort: 0},
		}
		assert.Equal(t, common.HA_MYSQL_DEFAULT_RW_PORT, mgr.PrimaryPort(haConfig))
	})
}

// ---------------------------------------------------------------------------
// MySQLHAConnectivityManager — AdditionalServices
// ---------------------------------------------------------------------------

func TestMySQLHAConnectivityManager_AdditionalServices(t *testing.T) {
	t.Parallel()

	mgr := &MySQLHAConnectivityManager{}

	t.Run("returns -ro-svc on port 3306 when router is not deployed and MySQL config is nil", func(t *testing.T) {
		haConfig := &ndbv1alpha1.InstanceHAConfig{MySQL: nil}
		svcs := mgr.AdditionalServices(haConfig)
		assert.Len(t, svcs, 1)
		assert.Equal(t, "-ro-svc", svcs[0].NameSuffix)
		assert.Equal(t, common.HA_MYSQL_DEFAULT_LISTENER_PORT, svcs[0].Port)
	})

	t.Run("returns -ro-svc on port 3306 when DeployMySQLRouter is false", func(t *testing.T) {
		haConfig := &ndbv1alpha1.InstanceHAConfig{
			MySQL: &ndbv1alpha1.MySQLHAConfig{DeployMySQLRouter: false},
		}
		svcs := mgr.AdditionalServices(haConfig)
		assert.Len(t, svcs, 1)
		assert.Equal(t, common.HA_MYSQL_DEFAULT_LISTENER_PORT, svcs[0].Port)
	})

	t.Run("returns -ro-svc on configured RouterROPort when router is deployed", func(t *testing.T) {
		haConfig := &ndbv1alpha1.InstanceHAConfig{
			MySQL: &ndbv1alpha1.MySQLHAConfig{DeployMySQLRouter: true, RouterROPort: 6500},
		}
		svcs := mgr.AdditionalServices(haConfig)
		assert.Len(t, svcs, 1)
		assert.Equal(t, "-ro-svc", svcs[0].NameSuffix)
		assert.Equal(t, int32(6500), svcs[0].Port)
	})

	t.Run("returns -ro-svc on default RO port when router is deployed and RouterROPort is zero", func(t *testing.T) {
		haConfig := &ndbv1alpha1.InstanceHAConfig{
			MySQL: &ndbv1alpha1.MySQLHAConfig{DeployMySQLRouter: true, RouterROPort: 0},
		}
		svcs := mgr.AdditionalServices(haConfig)
		assert.Len(t, svcs, 1)
		assert.Equal(t, common.HA_MYSQL_DEFAULT_RO_PORT, svcs[0].Port)
	})
}

// ---------------------------------------------------------------------------
// MySQLHAIPResolver — collectMasterIP
// ---------------------------------------------------------------------------

func TestMySQLHAIPResolver_collectMasterIP(t *testing.T) {
	t.Parallel()

	r := &MySQLHAIPResolver{}

	t.Run("returns Master node IP", func(t *testing.T) {
		nodes := []ndb_api.DatabaseNode{
			masterDBNode("10.0.0.1"),
			replicaDBNode("10.0.0.2"),
			replicaDBNode("10.0.0.3"),
		}
		ips := r.collectMasterIP(nodes)
		assert.Equal(t, []string{"10.0.0.1"}, ips)
	})

	t.Run("returns nil when no Master node is present", func(t *testing.T) {
		nodes := []ndb_api.DatabaseNode{
			replicaDBNode("10.0.0.2"),
			replicaDBNode("10.0.0.3"),
		}
		ips := r.collectMasterIP(nodes)
		assert.Nil(t, ips)
	})

	t.Run("returns nil when Master node has no IP", func(t *testing.T) {
		nodes := []ndb_api.DatabaseNode{
			{
				Properties: []ndb_api.Property{{Name: "role", Value: common.HA_NODE_ROLE_MASTER}},
				DbServer:   ndb_api.DatabaseServer{IPAddresses: []string{}},
			},
		}
		ips := r.collectMasterIP(nodes)
		assert.Nil(t, ips)
	})

	t.Run("returns nil for empty node list", func(t *testing.T) {
		ips := r.collectMasterIP([]ndb_api.DatabaseNode{})
		assert.Nil(t, ips)
	})
}

// ---------------------------------------------------------------------------
// MySQLHAIPResolver — ResolveIPs
// ---------------------------------------------------------------------------

// dbNodeWithServerId returns a database node with a DatabaseServerId set,
// matching the real-world NDB response where every DB VM has a server ID.
func dbNodeWithServerId(ip, serverId string, role string) ndb_api.DatabaseNode {
	return ndb_api.DatabaseNode{
		DatabaseServerId: serverId,
		Properties:       []ndb_api.Property{{Name: "role", Value: role}},
		DbServer:         ndb_api.DatabaseServer{IPAddresses: []string{ip}},
	}
}

// mockHTTPJSONResponse builds a minimal *http.Response whose body is the given JSON string.
func mockHTTPJSONResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestMySQLHAIPResolver_ResolveIPs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := &MySQLHAIPResolver{}

	// ── DatabaseServerId empty: DPC call skipped, fall back to Master IP ──────────

	t.Run("skips DPC call and returns Master IP when DatabaseServerId is empty", func(t *testing.T) {
		// Passing nil NDB client intentionally — if a DPC call were made this would panic.
		db := ndb_api.DatabaseResponse{
			DatabaseNodes: []ndb_api.DatabaseNode{
				masterDBNode("10.0.0.1"),
				replicaDBNode("10.0.0.2"),
				replicaDBNode("10.0.0.3"),
			},
		}
		ips, err := r.ResolveIPs(ctx, nil, db)
		assert.NoError(t, err)
		assert.Equal(t, []string{"10.0.0.1"}, ips)
	})

	t.Run("skips DPC call and returns nil when DatabaseServerId is empty and no Master node", func(t *testing.T) {
		db := ndb_api.DatabaseResponse{
			DatabaseNodes: []ndb_api.DatabaseNode{
				replicaDBNode("10.0.0.2"),
				replicaDBNode("10.0.0.3"),
			},
		}
		ips, err := r.ResolveIPs(ctx, nil, db)
		assert.NoError(t, err)
		assert.Nil(t, ips)
	})

	t.Run("skips DPC call and returns Master IP when router in databaseNodes has no IPs and DatabaseServerId is empty", func(t *testing.T) {
		// Router node is present but has no IPs and no server ID — DPC cannot be used.
		// Falls back to the Master VM IP as the best available connection point.
		db := ndb_api.DatabaseResponse{
			DatabaseNodes: []ndb_api.DatabaseNode{
				{
					DatabaseServerId: "",
					Properties:       []ndb_api.Property{{Name: "node_type", Value: common.HA_NODE_TYPE_MYSQLROUTER}},
					DbServer:         ndb_api.DatabaseServer{IPAddresses: []string{}},
				},
				masterDBNode("10.0.0.1"),
			},
		}
		ips, err := r.ResolveIPs(ctx, nil, db)
		assert.NoError(t, err)
		assert.Equal(t, []string{"10.0.0.1"}, ips)
	})

	// ── DPC-based lookup: the main path in production ─────────────────────────────
	// NDB does not include MySQL Router VMs in databaseNodes[]. The resolver calls
	// GetMySQLRouterIPsFromDPC which uses GET /dpcs/{id} to read router_info.

	t.Run("router-enabled: DPC lookup resolves router IPs when routers absent from databaseNodes", func(t *testing.T) {
		db := ndb_api.DatabaseResponse{
			DatabaseNodes: []ndb_api.DatabaseNode{
				dbNodeWithServerId("10.0.0.1", "server-id-db1", common.HA_NODE_ROLE_MASTER),
				dbNodeWithServerId("10.0.0.2", "server-id-db2", common.HA_NODE_ROLE_REPLICA),
			},
		}
		ndbClient := &mockNDBClient{}
		// GetDatabaseServer(server-id-db1) → returns dbserverClusterId
		ndbClient.On("NewRequest", http.MethodGet, "dbservers/server-id-db1", nil).Return(&http.Request{}, nil)
		ndbClient.On("Do", mock.Anything).Return(mockHTTPJSONResponse(
			`{"id":"server-id-db1","dbserverClusterId":"cluster-id-123"}`), nil).Once()
		// GetDPC(cluster-id-123) → DPC response with router_info containing two router VMs
		ndbClient.On("NewRequest", http.MethodGet, "dpcs/cluster-id-123", nil).Return(&http.Request{}, nil)
		ndbClient.On("Do", mock.Anything).Return(mockHTTPJSONResponse(
			`{"id":"cluster-id-123","info":{"info":{"cluster_info":{"router_info":{"router_node_list":[`+
				`{"host_id":"r1","host_ip":"10.0.1.1","host_name":"mysqlrouter-1"},`+
				`{"host_id":"r2","host_ip":"10.0.1.2","host_name":"mysqlrouter-2"}]}}}}}`), nil).Once()

		ips, err := r.ResolveIPs(ctx, ndbClient, db)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"10.0.1.1", "10.0.1.2"}, ips)
		ndbClient.AssertExpectations(t)
	})

	t.Run("router-disabled: DPC lookup finds no routers and falls back to Master IP", func(t *testing.T) {
		db := ndb_api.DatabaseResponse{
			DatabaseNodes: []ndb_api.DatabaseNode{
				dbNodeWithServerId("10.0.0.1", "server-id-db1", common.HA_NODE_ROLE_MASTER),
				dbNodeWithServerId("10.0.0.2", "server-id-db2", common.HA_NODE_ROLE_REPLICA),
			},
		}
		ndbClient := &mockNDBClient{}
		// GetDatabaseServer → cluster ID
		ndbClient.On("NewRequest", http.MethodGet, "dbservers/server-id-db1", nil).Return(&http.Request{}, nil)
		ndbClient.On("Do", mock.Anything).Return(mockHTTPJSONResponse(
			`{"id":"server-id-db1","dbserverClusterId":"cluster-id-456"}`), nil).Once()
		// GetDPC(cluster-id-456) → DPC with empty router_node_list (router not deployed)
		ndbClient.On("NewRequest", http.MethodGet, "dpcs/cluster-id-456", nil).Return(&http.Request{}, nil)
		ndbClient.On("Do", mock.Anything).Return(mockHTTPJSONResponse(
			`{"id":"cluster-id-456","info":{"info":{"cluster_info":{"router_info":{"router_node_list":[]}}}}}`), nil).Once()

		ips, err := r.ResolveIPs(ctx, ndbClient, db)
		assert.NoError(t, err)
		assert.Equal(t, []string{"10.0.0.1"}, ips)
		ndbClient.AssertExpectations(t)
	})

	t.Run("propagates NDB client error from DPC lookup", func(t *testing.T) {
		db := ndb_api.DatabaseResponse{
			DatabaseNodes: []ndb_api.DatabaseNode{
				dbNodeWithServerId("10.0.0.1", "server-id-456", common.HA_NODE_ROLE_MASTER),
			},
		}
		ndbClient := &mockNDBClient{}
		ndbClient.On("NewRequest", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("connection refused"))

		ips, err := r.ResolveIPs(ctx, ndbClient, db)
		assert.Error(t, err)
		assert.Nil(t, ips)
		ndbClient.AssertExpectations(t)
	})
}
