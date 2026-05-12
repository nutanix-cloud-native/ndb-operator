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

import "encoding/json"

// DPCProxyNode represents a single HAProxy VM entry inside DPCProxyInfo.
type DPCProxyNode struct {
	HostName string `json:"host_name"`
	HostIP   string `json:"host_ip"`
}

// DPCProxyInfo holds the Virtual IP and individual HAProxy node details for the cluster.
// ClusterIP is the Virtual IP (VIP); it is empty when no VIP was provisioned.
type DPCProxyInfo struct {
	ClusterIP     string         `json:"cluster_ip"`
	ProxyNodeList []DPCProxyNode `json:"proxy_node_list"`
}

// DPCClusterNode represents a single database VM entry in the cluster node_list.
// Present for both Postgres HA (roles: "master", "async_replica", "sync_replica")
// and MySQL HA (roles: "Master", "Replica").
type DPCClusterNode struct {
	HostId   string `json:"host_id"`
	HostIP   string `json:"host_ip"`
	HostName string `json:"host_name"`
	Role     string `json:"role"`
}

// DPCRouterNode represents a single MySQL Router VM in the DPC router info.
type DPCRouterNode struct {
	HostId   string `json:"host_id"`
	HostIP   string `json:"host_ip"`
	HostName string `json:"host_name"`
}

// DPCRouterInfo holds MySQL Router VM details returned by GET /dpcs/{id} for MySQL HA clusters.
// RouterNodeList is empty when MySQL Router was not deployed (router-disabled instance).
type DPCRouterInfo struct {
	RouterNodeList []DPCRouterNode `json:"router_node_list"`
}

// DPCClusterInfo is the nested cluster topology returned by GET /dpcs/{id}.
// NodeList is present for both Postgres HA and MySQL HA and lists all database VMs with their roles.
type DPCClusterInfo struct {
	// Shared fields
	NodeList []DPCClusterNode `json:"node_list"`
	// Postgres HA fields
	EtcdPort           int          `json:"etcd_port"`
	PatroniClusterName string       `json:"patroni_cluster_name"`
	ProxyInfo          DPCProxyInfo `json:"proxy_info"`
	// MySQL HA fields
	RouterInfo DPCRouterInfo `json:"router_info"`
}

// DPCInnerInfo is the second-level info block: .info.info in the NDB response.
type DPCInnerInfo struct {
	ClusterInfo DPCClusterInfo `json:"cluster_info"`
}

// DPCInfo is the first-level info block: .info in the NDB response.
// NDB nests the actual cluster topology one level deeper under another "info" key,
// so the full path to the VIP is .info.info.cluster_info.proxy_info.cluster_ip.
type DPCInfo struct {
	Info DPCInnerInfo `json:"info"`
}

// DPCResponse is the response body for GET /dpcs/{id}.
// NDB sometimes returns the `info` field as a JSON-encoded string rather than a nested object,
// so Info is captured as RawMessage and decoded by GetDPCInfo().
type DPCResponse struct {
	Id   string          `json:"id"`
	Name string          `json:"name"`
	Info json.RawMessage `json:"info"`
}

// GetDPCInfo decodes the Info field regardless of whether NDB returned it as a
// nested JSON object or as a JSON-encoded string (double-serialised).
func (r *DPCResponse) GetDPCInfo() (DPCInfo, error) {
	var info DPCInfo

	// Try direct object decode first.
	if err := json.Unmarshal(r.Info, &info); err == nil && info.Info.ClusterInfo.ProxyInfo.ClusterIP != "" {
		return info, nil
	}

	// NDB may serialise info as a JSON string — unquote and try again.
	var infoStr string
	if err := json.Unmarshal(r.Info, &infoStr); err == nil {
		if err2 := json.Unmarshal([]byte(infoStr), &info); err2 == nil {
			return info, nil
		}
	}

	// Return whatever we got from the first attempt (may be partially populated).
	json.Unmarshal(r.Info, &info) //nolint:errcheck
	return info, nil
}

// DPCDBServersOptions controls how the underlying DB server VMs are handled
// when a DBServerCluster (DPC) is deprovisioned.
type DPCDBServersOptions struct {
	Delete            bool `json:"delete"`
	DeleteVgs         bool `json:"deleteVgs"`
	DeleteVmSnapshots bool `json:"deleteVmSnapshots"`
}

// DPCDeprovisionRequest is the request body for DELETE /dpcs/{id}.
// This is the correct API for removing an HA DBServerCluster and all its member VMs.
type DPCDeprovisionRequest struct {
	Delete     bool                `json:"delete"`
	Remove     bool                `json:"remove"`
	SoftRemove bool                `json:"softRemove"`
	Forced     bool                `json:"forced"`
	DBServers  DPCDBServersOptions `json:"dbservers"`
}
