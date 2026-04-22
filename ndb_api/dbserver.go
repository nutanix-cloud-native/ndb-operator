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

package ndb_api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Deprovisions a database server vm given a server id
// Returns the task info summary response for the operation
func DeprovisionDatabaseServer(ctx context.Context, ndbClient ndb_client.NDBClientHTTPInterface, id string, req *DatabaseServerDeprovisionRequest) (task *TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	if id == "" {
		err = fmt.Errorf("id is empty")
		log.Error(err, "no database server id provided")
		return
	}
	if _, err = sendRequest(ctx, ndbClient, http.MethodDelete, "dbservers/"+id, req, &task); err != nil {
		log.Error(err, "Error in DeprovisionDatabaseServer")
		return
	}
	return
}

// GetDatabaseServer fetches details of a single DB server by ID.
// Used to retrieve the dbserverClusterId for HA cluster membership lookups.
func GetDatabaseServer(ctx context.Context, ndbClient ndb_client.NDBClientHTTPInterface, id string) (server *DBServerDetails, err error) {
	log := ctrllog.FromContext(ctx)
	if id == "" {
		err = fmt.Errorf("id is empty")
		log.Error(err, "no database server id provided")
		return
	}
	if _, err = sendRequest(ctx, ndbClient, http.MethodGet, "dbservers/"+id, nil, &server); err != nil {
		log.Error(err, "Error in GetDatabaseServer")
		return
	}
	return
}

// GetAllDatabaseServers returns all DB servers registered with NDB.
func GetAllDatabaseServers(ctx context.Context, ndbClient ndb_client.NDBClientHTTPInterface) (servers []DBServerDetails, err error) {
	log := ctrllog.FromContext(ctx)
	if _, err = sendRequest(ctx, ndbClient, http.MethodGet, "dbservers", nil, &servers); err != nil {
		log.Error(err, "Error in GetAllDatabaseServers")
	}
	return
}

// GetDPC fetches full DPC (DBServerCluster) details via GET /dpcs/{id}.
// The response includes the Virtual IP (if provisioned) and individual HAProxy node IPs.
func GetDPC(ctx context.Context, ndbClient ndb_client.NDBClientHTTPInterface, dpcId string) (dpc *DPCResponse, err error) {
	log := ctrllog.FromContext(ctx)
	if dpcId == "" {
		err = fmt.Errorf("dpcId is empty")
		log.Error(err, "no DBServerCluster id provided")
		return
	}
	if _, err = sendRequest(ctx, ndbClient, http.MethodGet, "dpcs/"+dpcId, nil, &dpc); err != nil {
		log.Error(err, "Error in GetDPC", "dpcId", dpcId)
	}
	return
}

// GetHAProxyIPsForCluster resolves the connection IPs for an HA cluster.
// It calls GET /dpcs/{clusterId} and checks proxy_info:
//   - If cluster_ip is set: the VIP was provisioned → returns [vip] (single IP, all traffic routes through it)
//   - If cluster_ip is empty: no VIP → returns individual HAProxy node IPs from proxy_node_list
//
// This is needed because NDB does not include HAProxy nodes in databaseNodes[] on GET /databases.
func GetHAProxyIPsForCluster(ctx context.Context, ndbClient ndb_client.NDBClientHTTPInterface, oneServerId string) (ips []string, err error) {
	log := ctrllog.FromContext(ctx)

	server, err := GetDatabaseServer(ctx, ndbClient, oneServerId)
	if err != nil {
		return nil, fmt.Errorf("GetHAProxyIPsForCluster: could not fetch server %s: %w", oneServerId, err)
	}
	if server.DbserverClusterId == "" {
		log.Info("Server has no dbserverClusterId; not an HA cluster", "id", oneServerId)
		return
	}

	dpc, err := GetDPC(ctx, ndbClient, server.DbserverClusterId)
	if err != nil {
		return nil, fmt.Errorf("GetHAProxyIPsForCluster: could not fetch DPC %s: %w", server.DbserverClusterId, err)
	}

	dpcInfo, err := dpc.GetDPCInfo()
	if err != nil {
		return nil, fmt.Errorf("GetHAProxyIPsForCluster: could not parse DPC info: %w", err)
	}
	log.Info("DPC proxy info parsed", "clusterIP", dpcInfo.Info.ClusterInfo.ProxyInfo.ClusterIP, "proxyNodeCount", len(dpcInfo.Info.ClusterInfo.ProxyInfo.ProxyNodeList), "rawInfoLength", len(dpc.Info))

	proxyInfo := dpcInfo.Info.ClusterInfo.ProxyInfo

	// VIP takes priority — a single IP handles both write and read traffic routing.
	if proxyInfo.ClusterIP != "" {
		log.Info("HA cluster has Virtual IP; using VIP", "vip", proxyInfo.ClusterIP, "dpcId", server.DbserverClusterId)
		return []string{proxyInfo.ClusterIP}, nil
	}

	// No VIP — use individual HAProxy node IPs from the DPC proxy node list.
	for _, node := range proxyInfo.ProxyNodeList {
		if node.HostIP != "" {
			log.Info("HA cluster HAProxy node", "name", node.HostName, "ip", node.HostIP)
			ips = append(ips, node.HostIP)
		}
	}
	if len(ips) == 0 {
		log.Info("No proxy IPs found in DPC proxy_node_list; falling back to dbservers name filter", "dpcId", server.DbserverClusterId)
		// Last-resort fallback: list all servers and filter by name.
		allServers, listErr := GetAllDatabaseServers(ctx, ndbClient)
		if listErr != nil {
			return nil, fmt.Errorf("GetHAProxyIPsForCluster: fallback list failed: %w", listErr)
		}
		for _, s := range allServers {
			if s.DbserverClusterId == server.DbserverClusterId &&
				strings.Contains(strings.ToLower(s.Name), common.HA_NODE_TYPE_HAPROXY) {
				ips = append(ips, s.IPAddresses...)
			}
		}
	}
	return
}

// DeprovisionDPC deletes an entire HA DBServerCluster (DPC) via DELETE /dpcs/{id}.
// This is the correct NDB API for removing HA instances — NDB blocks individual
// /dbservers/{id} deletes when the server belongs to a cluster.
// Returns the NDB operation ID so callers can poll for completion.
func DeprovisionDPC(ctx context.Context, ndbClient ndb_client.NDBClientHTTPInterface, dpcId string, req *DPCDeprovisionRequest) (operationId string, err error) {
	log := ctrllog.FromContext(ctx)
	if dpcId == "" {
		err = fmt.Errorf("dpcId is empty")
		log.Error(err, "no DBServerCluster id provided")
		return
	}
	var task *TaskInfoSummaryResponse
	if _, err = sendRequest(ctx, ndbClient, http.MethodDelete, "dpcs/"+dpcId, req, &task); err != nil {
		log.Error(err, "Error in DeprovisionDPC", "dpcId", dpcId)
		return
	}
	if task != nil {
		operationId = task.OperationId
	}
	return
}

// DeprovisionHADatabaseServers resolves the DBServerCluster ID from oneServerId, then
// deletes the entire cluster via DELETE /dpcs/{clusterId}.
// Returns a slice containing the single NDB operation ID so callers can poll for completion.
func DeprovisionHADatabaseServers(ctx context.Context, ndbClient ndb_client.NDBClientHTTPInterface, oneServerId string) (operationIds []string, err error) {
	log := ctrllog.FromContext(ctx)

	server, err := GetDatabaseServer(ctx, ndbClient, oneServerId)
	if err != nil {
		return nil, fmt.Errorf("could not fetch DB server %s: %w", oneServerId, err)
	}

	clusterId := server.DbserverClusterId
	if clusterId == "" {
		log.Info("DB server has no cluster ID; falling back to single-server delete", "id", oneServerId)
		siReq := GenerateDeprovisionDatabaseServerRequest()
		task, delErr := DeprovisionDatabaseServer(ctx, ndbClient, oneServerId, siReq)
		if delErr != nil {
			return nil, delErr
		}
		if task != nil && task.OperationId != "" {
			operationIds = []string{task.OperationId}
		}
		return
	}

	log.Info("Deprovisioning HA DBServerCluster via /dpcs endpoint", "dpcId", clusterId)
	dpcReq := GenerateDeprovisionDPCRequest()
	opId, delErr := DeprovisionDPC(ctx, ndbClient, clusterId, dpcReq)
	if delErr != nil {
		return nil, delErr
	}
	if opId != "" {
		operationIds = []string{opId}
	}
	return
}
