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
	"strings"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
)

// HAServiceSpec describes an extra Kubernetes service+endpoint pair to create for an HA database.
type HAServiceSpec struct {
	// NameSuffix is appended to the Database CR name to form the Service/Endpoint name.
	NameSuffix string
	// Port is the target port for this service.
	Port int32
}

// HAConnectivityManager defines engine-specific connectivity requirements for HA databases.
// Each engine that supports HA provides its own implementation.
// To add HA connectivity support for a new engine, implement this interface and register
// it in haConnectivityManagers.
type HAConnectivityManager interface {
	// PrimaryPort returns the port for the main -svc (e.g. HAProxy write port for Postgres).
	PrimaryPort(haConfig *ndbv1alpha1.InstanceHAConfig) int32
	// AdditionalServices returns specs for any extra services beyond the primary -svc.
	// e.g. Postgres HA returns a single -ro-svc for read replicas.
	AdditionalServices(haConfig *ndbv1alpha1.InstanceHAConfig) []HAServiceSpec
}

// haConnectivityManagers maps a database engine type to its HAConnectivityManager.
// To add HA connectivity support for a new engine, register its manager here.
var haConnectivityManagers = map[string]HAConnectivityManager{
	common.DATABASE_TYPE_POSTGRES: &PostgresHAConnectivityManager{},
	common.DATABASE_TYPE_MYSQL:    &MySQLHAConnectivityManager{},
	common.DATABASE_TYPE_MONGODB:  &MongoHAConnectivityManager{},
}

// HAIPResolver resolves the connection IP(s) for an HA database as reported by the NDB API.
// Each engine that supports HA provides its own implementation.
// To add IP resolution support for a new engine, implement this interface and register
// it in haIPResolvers.
type HAIPResolver interface {
	// ResolveIPs returns the connection IPs for an HA database.
	// For Postgres, these are the HAProxy VM IPs.
	// Returns an empty slice when no IPs can be resolved.
	ResolveIPs(ctx context.Context, ndbClient ndb_client.NDBClientHTTPInterface, db ndb_api.DatabaseResponse) ([]string, error)
}

// haIPResolvers maps a database engine type (DATABASE_TYPE_*) to its HAIPResolver.
// To add HA IP resolution for a new engine, register its resolver here.
var haIPResolvers = map[string]HAIPResolver{
	common.DATABASE_TYPE_POSTGRES: &PostgresHAIPResolver{},
	common.DATABASE_TYPE_MYSQL:    &MySQLHAIPResolver{},
	common.DATABASE_TYPE_MONGODB:  &MongoHAIPResolver{},
}

// PostgresHAIPResolver resolves HAProxy IPs for a Postgres HA database.
// It first scans databaseNodes[] for nodes whose properties or name identify them as HAProxy,
// then falls back to querying the DPC (DBServerCluster) endpoint if none are found there.
type PostgresHAIPResolver struct{}

func (r *PostgresHAIPResolver) ResolveIPs(ctx context.Context, ndbClient ndb_client.NDBClientHTTPInterface, db ndb_api.DatabaseResponse) ([]string, error) {
	// First try to find HAProxy nodes from the databaseNodes list (properties or name match).
	// NDB does not include HAProxy nodes in databaseNodes[], so this will usually return empty.
	ips := r.collectHAProxyIPs(db.DatabaseNodes)
	if len(ips) > 0 {
		return ips, nil
	}

	// Fallback: query the DPC (DBServerCluster) endpoint to find HAProxy IPs by cluster membership.
	if db.DatabaseNodes[0].DatabaseServerId == "" {
		return nil, nil
	}
	return ndb_api.GetHAProxyIPsForCluster(ctx, ndbClient, db.DatabaseNodes[0].DatabaseServerId)
}

// collectHAProxyIPs returns the IP of each HAProxy node in the list.
// A node is identified as HAProxy by a {node_type: "haproxy"} property,
// or as a fallback by its server name containing "haproxy" (NDB naming convention).
func (r *PostgresHAIPResolver) collectHAProxyIPs(nodes []ndb_api.DatabaseNode) []string {
	var ips []string
	for _, node := range nodes {
		isHAProxy := false
		for _, prop := range node.Properties {
			if prop.Name == "node_type" && prop.Value == common.HA_NODE_TYPE_HAPROXY {
				isHAProxy = true
				break
			}
		}
		if !isHAProxy && strings.Contains(strings.ToLower(node.DbServer.Name), common.HA_NODE_TYPE_HAPROXY) {
			isHAProxy = true
		}
		if isHAProxy && len(node.DbServer.IPAddresses) > 0 {
			ips = append(ips, node.DbServer.IPAddresses[0])
		}
	}
	return ips
}

// PostgresHAConnectivityManager manages HAProxy-based connectivity for Postgres HA.
// Primary -svc routes to the HAProxy write port; -ro-svc routes to the read port.
type PostgresHAConnectivityManager struct{}

func (m *PostgresHAConnectivityManager) PrimaryPort(haConfig *ndbv1alpha1.InstanceHAConfig) int32 {
	if pg := haConfig.Postgres; pg != nil && pg.WritePort != 0 {
		return pg.WritePort
	}
	return common.HA_PROXY_DEFAULT_WRITE_PORT
}

func (m *PostgresHAConnectivityManager) AdditionalServices(haConfig *ndbv1alpha1.InstanceHAConfig) []HAServiceSpec {
	readPort := common.HA_PROXY_DEFAULT_READ_PORT
	if pg := haConfig.Postgres; pg != nil && pg.ReadPort != 0 {
		readPort = pg.ReadPort
	}
	return []HAServiceSpec{
		{NameSuffix: "-ro-svc", Port: readPort},
	}
}

// MySQLHAConnectivityManager manages MySQL Router-based connectivity for MySQL HA.

type MySQLHAConnectivityManager struct{}

func (m *MySQLHAConnectivityManager) PrimaryPort(haConfig *ndbv1alpha1.InstanceHAConfig) int32 {
	my := haConfig.MySQL
	if my == nil || !my.DeployMySQLRouter {
		return common.HA_MYSQL_DEFAULT_LISTENER_PORT
	}
	if my.RouterRWPort != 0 {
		return my.RouterRWPort
	}
	return common.HA_MYSQL_DEFAULT_RW_PORT
}

func (m *MySQLHAConnectivityManager) AdditionalServices(haConfig *ndbv1alpha1.InstanceHAConfig) []HAServiceSpec {
	my := haConfig.MySQL
	if my == nil || !my.DeployMySQLRouter {
		return []HAServiceSpec{
			{NameSuffix: "-ro-svc", Port: common.HA_MYSQL_DEFAULT_LISTENER_PORT},
		}
	}
	roPort := common.HA_MYSQL_DEFAULT_RO_PORT
	if my.RouterROPort != 0 {
		roPort = my.RouterROPort
	}
	return []HAServiceSpec{
		{NameSuffix: "-ro-svc", Port: roPort},
	}
}

// MySQLHAIPResolver resolves connection IPs for a MySQL HA database.
type MySQLHAIPResolver struct{}

func (r *MySQLHAIPResolver) ResolveIPs(ctx context.Context, ndbClient ndb_client.NDBClientHTTPInterface, db ndb_api.DatabaseResponse) ([]string, error) {

	// /dpcs endpoint lists routers explicitly under router_info.router_node_list. When routers were deployed
	if len(db.DatabaseNodes) == 0 || db.DatabaseNodes[0].DatabaseServerId == "" {
		return r.collectMasterIP(db.DatabaseNodes), nil
	}
	routerIPs, err := ndb_api.GetMySQLRouterIPsFromDPC(ctx, ndbClient, db.DatabaseNodes[0].DatabaseServerId)
	if err != nil {
		return nil, err
	}
	if len(routerIPs) > 0 {
		return routerIPs, nil
	}

	// No routers in the cluster — router-disabled instance. Return Master VM IP.
	return r.collectMasterIP(db.DatabaseNodes), nil
}

// collectMasterIP returns the IP of the Master database node.
func (r *MySQLHAIPResolver) collectMasterIP(nodes []ndb_api.DatabaseNode) []string {
	for _, node := range nodes {
		for _, prop := range node.Properties {
			if prop.Name == "role" && prop.Value == common.HA_NODE_ROLE_MASTER {
				if len(node.DbServer.IPAddresses) > 0 {
					return []string{node.DbServer.IPAddresses[0]}
				}
			}
		}
	}
	return nil
}

// MongoHAIPResolver resolves connection IPs for a MongoDB HA (Replica Set) database.
// It returns the IPs of all data-bearing nodes (Primary + Secondaries), excluding Arbiters.
// These IPs are joined and stored in database.Status.IPAddress, then used by
// setupMongoConnSecret to build the full MongoDB connection URI.
type MongoHAIPResolver struct{}

func (r *MongoHAIPResolver) ResolveIPs(_ context.Context, _ ndb_client.NDBClientHTTPInterface, db ndb_api.DatabaseResponse) ([]string, error) {
	var ips []string
	for _, node := range db.DatabaseNodes {
		isArbiter := false
		for _, prop := range node.Properties {
			if prop.Name == "role" && prop.Value == common.HA_NODE_ROLE_MONGO_ARBITER {
				isArbiter = true
				break
			}
		}
		if !isArbiter && len(node.DbServer.IPAddresses) > 0 {
			ips = append(ips, node.DbServer.IPAddresses[0])
		}
	}
	return ips, nil
}

// MongoHAConnectivityManager manages connectivity for MongoDB HA (Replica Set) databases.
// Unlike Postgres/MySQL, MongoDB HA does NOT use K8s Services — the operator instead creates
// a <database-name>-db-uri Secret with the full MongoDB connection URI so that smart MongoDB
// drivers can connect directly to each node without K8s Service NAT interfering with SDAM.
type MongoHAConnectivityManager struct{}

func (m *MongoHAConnectivityManager) PrimaryPort(haConfig *ndbv1alpha1.InstanceHAConfig) int32 {
	if mg := haConfig.MongoDB; mg != nil && mg.ListenerPort != 0 {
		return mg.ListenerPort
	}
	return common.HA_MONGO_DEFAULT_LISTENER_PORT
}

// AdditionalServices returns nil — MongoDB HA creates no K8s Services.
// Connectivity is handled via the <database-name>-db-uri Secret instead.
func (m *MongoHAConnectivityManager) AdditionalServices(_ *ndbv1alpha1.InstanceHAConfig) []HAServiceSpec {
	return nil
}
