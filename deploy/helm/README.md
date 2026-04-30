# Nutanix Database Service Operator for Kubernetes

The NDB operator automates and simplifies database administration, provisioning, and life-cycle management of NDB on Kubernetes.

NDB operator supports these functionalities:

1. Provisioning and deprovisioning a single instance postgres, mssql, sql server, and mongodb database with or without time machine.
2. Cloning support for the above database engines
3. Provisioning and deprovisioning Postgres High Availability (HA) instances.
4. Creation of a service for the applications to consume the database within Kubernetes.

---

## Pre-requisites

1. [Install](https://portal.nutanix.com/page/documents/details?targetId=Nutanix-NDB-User-Guide-v2_8:Nutanix-NDB-User-Guide-v2_8) NDB 2.8.
2. [Install](https://helm.sh/docs/intro/install/) Helm v3.0.0.
3. [Install](https://kubernetes.io/docs/setup/) a Kubernetes cluster.
4. [Install](https://cert-manager.io/docs/installation/#getting-started) cert-manager. Ensure that the cert-manager resouces are up and running successfully before installing the NDB operator.

## Installation and Running on the cluster

### Method 1: Install from Helm Repository (Recommended)

Add the Nutanix Helm repository:

```sh
helm repo add nutanix https://nutanix.github.io/helm/
helm repo update
```

Install the latest version:

```sh
helm install ndb-operator nutanix/ndb-operator \
  --namespace ndb-operator-system \
  --create-namespace
```

Install a specific version:

```sh
helm install ndb-operator nutanix/ndb-operator \
  --version 0.5.7 \
  --namespace ndb-operator-system \
  --create-namespace
```

List available versions:

```sh
helm search repo nutanix/ndb-operator --versions
```

### Method 2: Install from OCI Registry (GHCR)

Install the latest version:

```sh
helm install ndb-operator oci://ghcr.io/nutanix-cloud-native/chart/ndb-operator \
  --namespace ndb-operator-system \
  --create-namespace
```

Install a specific version:

```sh
helm install ndb-operator oci://ghcr.io/nutanix-cloud-native/chart/ndb-operator \
  --version 0.5.7 \
  --namespace ndb-operator-system \
  --create-namespace
```

List available versions:

```sh
# Install oras to list OCI registry tags
brew install oras

# List all available versions
oras repo tags ghcr.io/nutanix-cloud-native/chart/ndb-operator
```

### Verify Installation

Check the installation status:

```sh
# Check Helm release
helm list -n ndb-operator-system

# Check pods
kubectl get pods -n ndb-operator-system

# Check CRDs
kubectl get crds | grep ndb.nutanix.com
```

### Upgrading

To upgrade to a newer version:

```sh
# Using Helm repository
helm repo update
helm upgrade ndb-operator nutanix/ndb-operator \
  --namespace ndb-operator-system

# Using OCI registry
helm upgrade ndb-operator oci://ghcr.io/nutanix-cloud-native/chart/ndb-operator \
  --version 0.5.7 \
  --namespace ndb-operator-system
```

### Uninstalling

To uninstall the operator:

```sh
helm uninstall ndb-operator --namespace ndb-operator-system
```

## Usage

For the complete operator guide (including migration notes and development), see the [main repository README](https://github.com/nutanix-cloud-native/ndb-operator/blob/main/README.md).

**NDBServer and credentials:** The operator uses two custom resources—**NDBServer** (cluster-scoped) and **Database** (namespaced). **NDBServer** is cluster-scoped so that admins can store the NDB API credential secret in a restricted namespace (e.g. `ndb-credentials`) and set `credentialSecretRef` to point to it. Developers who create **Database** resources only need to reference the NDBServer by **name** in `ndbRef` (e.g. `ndbRef: ndb`); they can list and use cluster-scoped NDBServers without needing access to the secret's namespace.

### Create secrets to be used by the NDBServer and Database resources using the manifest:

- **NDB API credential secret:** Create this in a **restricted namespace** (e.g. `ndb-credentials`) so only admins need access. Create that namespace if it does not exist, then apply the secret there.
- **Database instance secret:** Create this in the same namespace where you will create the Database resource (e.g. your application namespace).

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: ndb-secret-name
  namespace: ndb-credentials   # use a restricted namespace; create it first
type: Opaque
stringData:
  username: username-for-ndb-server
  password: password-for-ndb-server
  ca_certificate: |
    -----BEGIN CERTIFICATE-----
    CA CERTIFICATE (ca_certificate is optional)
    -----END CERTIFICATE-----
---
apiVersion: v1
kind: Secret
metadata:
  name: db-instance-secret-name
  # no namespace, or set to the namespace where you will create the Database
type: Opaque
stringData:
  password: password-for-the-database-instance
  ssh_public_key: SSH-PUBLIC-KEY

```

Create the NDB credential namespace, then apply the secrets (the NDB secret YAML above includes `namespace: ndb-credentials`):

```sh
kubectl create namespace ndb-credentials
kubectl apply -f <path/to/secrets-manifest.yaml>
```

### Create the NDBServer resource. The manifest for NDBServer is described as follows:

**NDBServer is cluster-scoped.** Admins create the NDB API credential secret in a restricted namespace (e.g. `ndb-credentials`) and set `credentialSecretRef` to that secret. The NDBServer resource itself has no namespace; developers in any namespace can reference it by name.

```yaml
apiVersion: ndb.nutanix.com/v1alpha1
kind: NDBServer
metadata:
  labels:
    app.kubernetes.io/name: ndbserver
    app.kubernetes.io/instance: ndbserver
    app.kubernetes.io/part-of: ndb-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: ndb-operator
  name: ndb
  # no namespace: NDBServer is cluster-scoped
spec:
    # Reference to the secret that holds the credentials for NDB (username, password, ca_certificate).
    # Point to the restricted namespace where the secret was created; developers do not need access to this namespace.
    credentialSecretRef:
      name: ndb-secret-name
      namespace: ndb-credentials
    # NDB Server's API URL
    server: https://[NDB IP]:8443/era/v0.9
    # Set to true to skip SSL certificate validation, should be false if ca_certificate is provided in the credential secret.
    skipCertificateVerification: true

```

Create the NDBServer resource using:

```sh
kubectl apply -f <path/to/NDBServer-manifest.yaml>
```

### Create a Database Resource. A database can either be provisioned or cloned on NDB based on the inputs specified in the database manifest.

#### Using ConfigMap Defaults (Optional)

The NDB Operator supports using a ConfigMap to provide default values for database configurations. This allows administrators to pre-configure common settings, reducing the amount of configuration developers need to specify in their Database CRs.

Defaults are applied by the **mutating webhook** (defaulter) before the CR is validated and persisted. The controller receives the fully populated CR, so validation is always strict and there is no duplicate logic.

**How it works:**

- Set `defaultsConfigMapRef` in the Database spec to the name of a ConfigMap in the **same namespace** as the Database CR.
- The defaulter webhook fetches the ConfigMap, applies defaults to empty fields, then runs standard defaulting. Validation runs on the fully populated CR.
- If the ConfigMap does not exist or cannot be fetched, the webhook proceeds without ConfigMap defaults (logs a message) and uses standard defaults.
- If the ConfigMap exists but `data` is empty, no defaults are applied from it (same as having no keys).
- Omit `defaultsConfigMapRef` to use the traditional flow—no ConfigMap, no change in behavior.
- The operator’s ServiceAccount must be allowed to **get** and **list** ConfigMaps in namespaces where `Database` resources are created (the shipped RBAC includes this).

**Benefits:**

- Simplifies Database CR definitions
- Centralizes common configuration
- Easy to update defaults without modifying Database CRs

**Key precedence:** Explicitly set fields on the Database CR take precedence; the ConfigMap only fills **empty** fields. Engine-specific keys (e.g. `postgres.profiles.software.name`) are evaluated before generic keys (e.g. `profiles.software.name`). For clones, `clone.*` keys (e.g. `clone.timezone`, `clone.profiles.software.name`) are evaluated before the shared generic keys. **`size`** keys apply to **provisioning** only, not cloning.

**Supported ConfigMap Keys:**

| Key | Applies To | Description |
|-----|------------|-------------|
| `clusterName` | Provision, Clone | NDB cluster name |
| `timezone` | Provision, Clone | Database timezone |
| `size` | Provision | DB size in GB |
| `profiles.compute.name` | Provision, Clone | Compute profile |
| `profiles.network.name` | Provision, Clone | Network profile |
| `profiles.software.name` | Provision, Clone | Software profile |
| `profiles.dbParam.name` | Provision, Clone | DB parameter profile |
| `profiles.dbParamInstance.name` | Provision, Clone | DB param instance (MSSQL) |
| `timeMachine.sla` | Provision | SLA name |
| `timeMachine.dailySnapshotTime` | Provision | Daily snapshot time (hh:mm:ss) |
| `timeMachine.snapshotsPerDay` | Provision | Snapshots per day |
| `timeMachine.logCatchUpFrequency` | Provision | Log catch-up (minutes) |
| `timeMachine.weeklySnapshotDay` | Provision | Weekly snapshot day |
| `timeMachine.monthlySnapshotDay` | Provision | Monthly snapshot day |
| `timeMachine.quarterlySnapshotMonth` | Provision | Quarterly snapshot month |
| `postgres.size`, `mysql.size`, `mongodb.size`, `mssql.size` | Provision | Engine-specific size (GB); not used for clone |
| `postgres.profiles.*`, `mysql.profiles.*`, `mongodb.profiles.*`, `mssql.profiles.*` | Provision, Clone | Engine-specific profile defaults |
| `clone.clusterName`, `clone.timezone`, `clone.profiles.*` | Clone | Clone-specific keys (take precedence over generic keys for clones) |

If a key is present in the ConfigMap, it is applied when the corresponding CR field is empty. To rely on NDB OOB profile resolution for a given profile slot instead, omit that key from the ConfigMap (or set the profile explicitly on the CR).

**Quick Example:**

```yaml
# 1. Create a ConfigMap with defaults
apiVersion: v1
kind: ConfigMap
metadata:
  name: ndb-database-defaults
  namespace: default
data:
  clusterName: "production-cluster"
  timezone: "UTC"
  profiles.compute.name: "DEFAULT_OOB_COMPUTE"
  profiles.network.name: "DEFAULT_OOB_NETWORK"
  timeMachine.sla: "DEFAULT_OOB_BRASS_SLA"
  # PostgreSQL-specific defaults
  postgres.profiles.software.name: "POSTGRES_15.6_OOB"
  postgres.profiles.dbParam.name: "DEFAULT_POSTGRES_PARAMS"
  postgres.size: "10"
  # MySQL-specific defaults
  mysql.profiles.software.name: "MYSQL_8.0_OOB"
  mysql.profiles.dbParam.name: "DEFAULT_MYSQL_PARAMS"
  mysql.size: "10"

---
# 2. Create a minimal Database CR using the defaults
apiVersion: ndb.nutanix.com/v1alpha1
kind: Database
metadata:
  name: my-app-db
spec:
  ndbRef: ndb
  defaultsConfigMapRef: ndb-database-defaults  # Reference the ConfigMap
  isClone: false
  databaseInstance:
    type: postgres
    name: my-app-db
    databaseNames: ["appdb"]
    credentialSecret: db-instance-secret-name
    # All other fields (cluster, size, profiles, timeMachine) come from ConfigMap!
```

#### Using Database CRs traditional way (without configmap)

#### Provisioning manifest

```yaml
apiVersion: ndb.nutanix.com/v1alpha1
kind: Database
metadata:
  # This name that will be used within the kubernetes cluster
  name: db
spec:
  # Name of the cluster-scoped NDBServer (no namespace needed; developers reference by name only)
  ndbRef: ndb
  isClone: false
  # Database instance specific details (that is to be provisioned)
  databaseInstance:
    # Cluster Name or cluster ID where the Database has to be provisioned
    # Can be fetched from the GET /clusters endpoint
    clusterName: "Nutanix Cluster Name"         # Recommended: Use cluster name
    # clusterId: "Nutanix Cluster UUID"         # Alternative: Use cluster UUID
    # The database instance name on NDB
    name: "Database-Instance-Name"
    # The description of the database instance
    description: Database Description
    # Names of the databases on that instance
    databaseNames:
      - database_one
      - database_two
      - database_three
    # Credentials secret name for NDB installation
    # data: password, ssh_public_key
    credentialSecret: db-instance-secret-name
    size: 10
    timezone: "UTC"
    type: postgres

    # You can specify any (or none) of these types of profiles: compute, software, network, dbParam
    # If not specified, the corresponding Out-of-Box (OOB) profile will be used wherever applicable
    # Name is case-sensitive. ID is the UUID of the profile. Profile should be in the "READY" state
    # "id" & "name" are optional. If none provided, OOB may be resolved to any profile of that type
    profiles:
      compute:
        id: ""
        name: ""
      # A Software profile is a mandatory input for closed-source engines: SQL Server & Oracle
      software:
        name: ""
        id: ""
      network:
        id: ""
        name: ""
      dbParam:
        name: ""
        id: ""
      # Only applicable for MSSQL databases
      dbParamInstance:
        name: ""
        id: ""
    timeMachine:                        # Optional block, if removed the SLA defaults to NONE
      sla : "NAME OF THE SLA"
      dailySnapshotTime:   "12:34:56"   # Time for daily snapshot in hh:mm:ss format
      snapshotsPerDay:     4            # Number of snapshots per day
      logCatchUpFrequency: 90           # Frequency (in minutes)
      weeklySnapshotDay:   "WEDNESDAY"  # Day of the week for weekly snapshot
      monthlySnapshotDay:  24           # Day of the month for monthly snapshot
      quarterlySnapshotMonth: "Jan"     # Start month of the quarterly snapshot
    additionalArguments:                # Optional block, can specify additional arguments that are unique to database engines.
      listener_port: "8080"

```

#### Cloning manifest

```yaml
apiVersion: ndb.nutanix.com/v1alpha1
kind: Database
metadata:
  # This name that will be used within the kubernetes cluster
  name: db
spec:
  # Name of the cluster-scoped NDBServer (no namespace needed; developers reference by name only)
  ndbRef: ndb
  isClone: true
  # Clone specific details (that is to be provisioned)
  clone:
    # Type of the database to be cloned
    type: postgres
    # The clone instance name on NDB
    name: "Clone-Instance-Name"
    # The description of the clone instance
    description: Database Description
    # Cluster Name or Cluster id of the cluster where the Cloned Database has to be provisioned
    # Can be fetched from the GET /clusters endpoint
    clusterName: "Nutanix Cluster Name"         # Recommended: Use cluster name
    # clusterId: "Nutanix Cluster UUID"         # Alternative: Use cluster UUID
    # You can specify any (or none) of these types of profiles: compute, software, network, dbParam
    # If not specified, the corresponding Out-of-Box (OOB) profile will be used wherever applicable
    # Name is case-sensitive. ID is the UUID of the profile. Profile should be in the "READY" state
    # "id" & "name" are optional. If none provided, OOB may be resolved to any profile of that type
    profiles:
      compute:
        id: ""
        name: ""
      # A Software profile is a mandatory input for closed-source engines: SQL Server & Oracle
      software:
        name: ""
        id: ""
      network:
        id: ""
        name: ""
      dbParam:
        name: ""
        id: ""
      # Only applicable for MSSQL databases
      dbParamInstance:
        name: ""
        id: ""
    # Name of the secret with the
    # data: password, ssh_public_key
    credentialSecret: clone-instance-secret-name
    timezone: "UTC"
    
    # Name or ID of the database to clone from, can be fetched from NDB REST API Explorer
    sourceDatabaseName: "source-database-name"      # Recommended: Use database name
    # sourceDatabaseId: "source-database-uuid"      # Alternative: Use database UUID
    
    # Name or ID of the snapshot to clone from, can be fetched from NDB REST API Explorer
    # AUTO-SNAPSHOT: If both snapshotName and snapshotId are omitted, the operator will
    # automatically select the most recent snapshot from the source database.
    # This works with or without ConfigMap defaults.
    snapshotName: "snapshot-name"                   # Recommended: Use snapshot name
    # snapshotId: "snapshot-uuid"                   # Alternative: Use snapshot UUID
    # Omit both to auto-select latest snapshot
    
    additionalArguments:                        # Optional block, can specify additional arguments that are unique to database engines.
      expireInDays: 3

```

Create the Database resource:

```sh
kubectl apply -f <path/to/database-manifest.yaml>
```

#### Creating Postgres HA instance resource
```yaml
apiVersion: ndb.nutanix.com/v1alpha1
kind: Database
metadata:
  name: Postgres-HA-K8s-resource
  namespace: default
spec:
  ndbRef: ndbserver
  databaseInstance:
    name: "PGHA_instance_DB"
    description: "Postgres HA instance"
    # defaultsConfigMapRef: pgha-defaults   # injects timezone, profiles, timeMachine from ConfigMap
    type: postgres
    credentialSecret: pgha-db-secret
    size: 200
    clusterName: "<PE cluster name (as shown in NDB UI)>" # This is often the Primary PE cluster
    # ClusterID: "UUID of the PE cluster"
    databaseNames:
      - PGHA_instance
    timeMachine:
      name: "PGHA_TM"
      description: "TM for Postgres HA"
    haConfig:
      patroniClusterName: "pgha-patroni" # any desired name
      clusterName: "PGHA_cluster" # any desired name
      enableSynchronousMode: true # default is false, This is for data replication across DB nodes
      # provisionVirtualIP: true # default is false, keep this true if having stretched VLAN and need to provision VirtualIP
      # writePort: 5000   # defaults to 5000 if omitted
      # readPort: 5001    # defaults to 5001 if omitted
      nodes:
        - vmName: "PGHA_haproxy1" # any desired name
          nodeType: "haproxy"
          clusterName: "<PE cluster name (as shown in NDB UI)>"
          # clusterId: "UUID of the PE cluster"
        - vmName: "PGHA_haproxy2" # any desired name
          nodeType: "haproxy"
          clusterName: "<PE cluster name (as shown in NDB UI)>"
          # clusterId: "UUID of the PE cluster"
        - vmName: "PGHA_DB-1" # any desired name
          nodeType: "database"
          role: "Primary"
          clusterName: "<PE cluster name (as shown in NDB UI)>"
          # clusterId: "UUID of the PE cluster"
        - vmName: "PGHA_DB-2" # any desired name
          nodeType: "database"
          role: "Secondary"
          clusterName: "<PE cluster name (as shown in NDB UI)>"
          # clusterId: "UUID of the PE cluster"
        - vmName: "PGHA_DB-3" # any desired name
          nodeType: "database"
          role: "Secondary"
          clusterName: "<PE cluster name (as shown in NDB UI)>"
          # clusterId: "UUID of the PE cluster"
```

### Example Defaults configmap for HA instance
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: pgha-defaults
  namespace: default
data:
  timezone: "UTC"
  # postgres-prefixed keys take priority over unprefixed keys for postgres databases
  postgres.profiles.software.name: "POSTGRES_15.6_HA_ENABLED_ROCKY_LINUX_8_OOB"
  postgres.profiles.compute.name: "DEFAULT_HA_COMPUTE"
  postgres.profiles.network.name: "PGHA_VLAN"
  postgres.profiles.dbParam.name: "DEFAULT_POSTGRES_HA_PARAMS"
  postgres.timeMachine.sla: "DEFAULT_OOB_BRONZE_SLA"
  postgres.timeMachine.dailySnapshotTime: "10:00:00"
  postgres.timeMachine.snapshotsPerDay: "1"
  postgres.timeMachine.logCatchUpFrequency: "120"
  postgres.timeMachine.weeklySnapshotDay: "WEDNESDAY"
  postgres.timeMachine.monthlySnapshotDay: "8"
  postgres.timeMachine.quarterlySnapshotMonth: "Jan"
```

### Additional Arguments for Databases

Below are the various optional addtionalArguments you can specify along with examples of their corresponding values. Arguments that have defaults will be indicated.

Provisioning Additional Arguments: 

```yaml
# PostGres
additionalArguments:
  listener_port: "1111"                            # Default: "5432"

# MySQL
additionalArguments:
  listener_port: "1111"                            # Default: "3306" 

# MongoDB
additionalArguments:
  listener_port: "1111"                            # Default: "27017"
  log_size: "150"                                  # Default: "100"
  journal_size: "150"                              # Default: "100"

# MSSQL
additionalArguments:
  sql_user_name: "mazin"                           # Defualt: "sa".
  authentication_mode: "mixed"                     # Default: "windows". Options are "windows" or "mixed". Must specify sql_user.
  server_collation: "<server-collation>"           # Default: "SQL_Latin1_General_CP1_CI_AS".
  database_collation:  "<server-collation>"        # Default: "SQL_Latin1_General_CP1_CI_AS".
  dbParameterProfileIdInstance: "<id-instance>"    # Default: Fetched from profile.
  vm_dbserver_admin_password: "<admin-password>"   # Default: Fetched from database secret.
  sql_user_password:         "<sq-user-password>"  # NO Default. Must specify authentication_mode as "mixed".
  windows_domain_profile_id: <domain-profile-id>   # NO Default. Must specify vm_db_server_user.
  vm_db_server_user: <vm-db-server-use>            # NO Default. Must specify windows_domain_profile_id.
  vm_win_license_key: <licenseKey>                 # NO Default.
```

Cloning Additional Arguments: 

```yaml
MSSQL:
  windows_domain_profile_id   
  era_worker_service_user      
  sql_service_startup_account  
  vm_win_license_key           
  target_mountpoints_location  
  expireInDays                 
  expiryDateTimezone           
  deleteDatabase               
  refreshInDays                
  refreshTime                  
  refreshDateTimezone          

MongoDB:
  expireInDays                 
  expiryDateTimezone           
  deleteDatabase               
  refreshInDays                
  refreshTime                  
  refreshDateTimezone    

Postgres:
  expireInDays                 
  expiryDateTimezone           
  deleteDatabase               
  refreshInDays                
  refreshTime                  
  refreshDateTimezone  

MySQL:
  expireInDays                 
  expiryDateTimezone           
  deleteDatabase               
  refreshInDays                
  refreshTime                  
  refreshDateTimezone  
```

### Deleting the Database resource

To deregister the database and delete the VM run:

```sh
kubectl delete -f <path/to/database-manifest.yaml>
```

### Deleting the NDBServer resource

To deregister the database and delete the VM run:

```sh
kubectl delete -f <path/to/NDBServer-manifest.yaml>
```

---

## Uninstalling the Chart

To uninstall/delete the operator deployment/chart:

```console
helm uninstall ndb-operator -n ndb-operator
```

---

## Configuration

The following table lists the configurable parameters of the NDB operator chart and their default values.


| Parameter          | Description                                            | Default                                                |
| ------------------ | ------------------------------------------------------ | ------------------------------------------------------ |
| `replicaCount`     | Number of replicas of the NDB Operator controller pods | `1`                                                    |
| `image.repository` | Image for NDB Operator controller                      | `ghcr.io/nutanix-cloud-native/ndb-operator/controller` |
| `image.pullPolicy` | Image pullPolicy                                       | `IfNotPresent`                                         |
| `image.tag`        | Image tag                                              | `""`                                                   |
| `imagePullSecrets` | ImagePullSecrets list                                  | `[]`                                                   |
| `fullnameOverride` | To override the full name of the operator chart        | `""`                                                   |
| `resources`        | Configure resources for Cloud Provider Pod             | `refer to values.yaml`                                 |
| `nodeSelector`     | Configure nodeSelector for Cloud Provider Pod          | `refer to values.yaml`                                 |
| `tolerations`      | Configure tolerations for Cloud Provider Pod           | `refer to values.yaml`                                 |
| `affinity`         | Configure affinity for Cloud Provider Pod              | `refer to values.yaml`                                 |


### Configuration examples:

Install the operator in the `ndb-operator` namespace (add the `--create-namespace` flag if the namespace does not exist): 

```console
helm install ndb-operator nutanix/ndb-operator -n ndb-operator 
```

Individual configurations can be set by using `--set key=value[,key=value]` like:

```console
helm install ndb-operator nutanix/ndb-operator  --set replicaCount=2 
```

In the above command `replicaCount` refers to one of the variables defined in the values.yaml file. 

All the options can also be specified in a value.yaml file:

```console
helm install ndb-operator nutanix/ndb-operator -f value.yaml
```

---

## How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

A custom resource of the kind Database is created by the reconciler, followed by a Service and an Endpoint that maps to the IP address of the database instance provisioned. Application pods/deployments can use this service to interact with the databases provisioned on NDB through the native Kubernetes service. 

Pods can specify an initContainer to wait for the service (and hence the database instance) to get created before they start up.

```yaml
  initContainers:
  - name: init-db
    image: busybox:1.28
    command: ['sh', '-c', "until nslookup <<Database CR Name>>-svc.$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local; do echo waiting for database service; sleep 2; done"]
```

## Contributing

See the [contributing docs](https://github.com/nutanix-cloud-native/ndb-operator/blob/main/CONTRIBUTING.md).

## Support

### Community Plus

This code is developed in the open with input from the community through issues and PRs. A Nutanix engineering team serves as the maintainer. Documentation is available in the project repository.

Issues and enhancement requests can be submitted in the [Issues tab of this repository](https://github.com/nutanix-cloud-native/ndb-operator/issues). Please search for and review the existing open issues before submitting a new issue.

## License

Copyright 2022-2026 Nutanix, Inc.

The project is released under version 2.0 of the [Apache license](http://www.apache.org/licenses/LICENSE-2.0).