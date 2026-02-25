# Nutanix Database Service Operator for Kubernetes
The NDB operator brings automated and simplified database administration, provisioning, and life-cycle management to Kubernetes.

---

[![Go Report Card](https://goreportcard.com/badge/github.com/nutanix-cloud-native/ndb-operator)](https://goreportcard.com/report/github.com/nutanix-cloud-native/ndb-operator)
![CI](https://github.com/nutanix-cloud-native/ndb-operator/actions/workflows/build-dev.yaml/badge.svg)
![Release](https://github.com/nutanix-cloud-native/ndb-operator/actions/workflows/release.yaml/badge.svg)

[![release](https://img.shields.io/github/release-pre/nutanix-cloud-native/ndb-operator.svg)](https://github.com/nutanix-cloud-native/ndb-operator/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/nutanix-cloud-native/ndb-operator/blob/master/LICENSE)
![Proudly written in Golang](https://img.shields.io/badge/written%20in-Golang-92d1e7.svg)

---
## Installation / Deployment
### Pre-requisites
1. Access to an NDB Server.
2. A Kubernetes cluster to run against, which should have network connectivity to the NDB server. The operator will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).
3. The [operator-sdk installed](https://sdk.operatorframework.io/docs/installation/).
4. A clone of the source code ([this](https://github.com/nutanix-cloud-native/ndb-operator) repository).
5. Cert-manager (only when running in non OpenShift clusters). Follow the instructions [here](https://cert-manager.io/docs/installation/).

With the pre-requisites completed, the NDB Operator can be deployed in one of the following ways: 

### Outside Kubernetes
Runs the controller outside the Kubernetes cluster as a process, but installs the CRDs, services and RBAC entities within the Kubernetes cluster. Generally used while development (without running webhooks):
```sh
make install run
```

### Within Kubernetes 
Runs the controller pod, installs the CRDs, services and RBAC entities within the Kubernetes cluster. Used to run the operator from the container image defined in the Makefile. Make sure that the cert-manager is installed if not using OpenShift.

```sh
make deploy
```

### Using Helm Charts
The Helm charts for the NDB Operator project are available on artifacthub.io and can be installed by following the instructions [here](https://artifacthub.io/packages/helm/nutanix/ndb-operator?modal=install).

### On OpenShift
To deploy the operator from this repository on an OpenShift cluster, create a bundle and then install the operator via the operator-sdk.
```sh
# Export these environment variables to overwrite the variables set in the Makefile
export DOCKER_USERNAME=dockerhub-username
export VERSION=x.y.z
export IMG=docker.io/$DOCKER_USERNAME/ndb-operator:v$VERSION
export BUNDLE_IMG=docker.io/$DOCKER_USERNAME/ndb-operator-bundle:v$VERSION

# Build and push the container image to the container registry
make docker-build docker-push

# Build the bundle following the prompts for input, build and push the bundle image to the container registry
make bundle bundle-build bundle-push

# Install the operator (run on the OpenShift cluster)
operator-sdk run bundle $BUNDLE_IMG

NOTE: 
The container and bundle image creation steps can be skipped if existing images are present in the container registry.
```

---

## Usage

**NDBServer and credentials:** The operator uses two custom resources—**NDBServer** (cluster-scoped) and **Database** (namespaced). **NDBServer** is cluster-scoped so that admins can store the NDB API credential secret in a restricted namespace (e.g. `ndb-credentials`) and set `credentialSecretRef` to point to it. Developers who create **Database** resources only need to reference the NDBServer by **name** in `ndbRef` (e.g. `ndbRef: ndb`); they can list and use cluster-scoped NDBServers without needing access to the secret's namespace.

###  Create secrets to be used by the NDBServer and Database resources using the manifest:

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

###  Create the NDBServer resource. The manifest for NDBServer is described as follows:

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

**Benefits:**
- Simplifies Database CR definitions
- Centralizes common configuration
- Easy to update defaults without modifying Database CRs
- Values in Database CR always override ConfigMap defaults

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
  postgres.profiles.software.name: "POSTGRES_15.6_OOB"
  postgres.size: "10"

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
    snapshotName: "snapshot-name"                   # Recommended: Use snapshot name, or leave empty for latest
    # snapshotId: "snapshot-uuid"                   # Alternative: Use snapshot UUID
    
    additionalArguments:                        # Optional block, can specify additional arguments that are unique to database engines.
      expireInDays: 3

```

Create the Database resource:
```sh
kubectl apply -f <path/to/database-manifest.yaml>
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

## Upgrading from 0.5.2 to 0.5.3 (NDBServer: namespaced → cluster-scoped)

In **0.5.3**, **NDBServer** is **cluster-scoped** and uses **`credentialSecretRef`** (secret `name` + `namespace`) instead of **`credentialSecret`** (name only). There is no in-place conversion; you must perform a one-time migration before or as part of the upgrade.

**Important:** The Kubernetes API does not allow changing a CRD from namespaced to cluster-scoped while custom resources of that kind exist. You must delete all existing **NDBServer** resources (after backing them up) before installing the 0.5.3 CRD.

### Step 1: Back up existing NDBServer(s)

Record the name and spec of each NDBServer so you can recreate them. For example:

```bash
kubectl get ndbserver -A -o yaml > ndbserver-backup.yaml
```

Note which namespace each NDBServer was in and which secret name it used (`spec.credentialSecret`).

### Step 2: Create credentials namespace and secret

Create a dedicated namespace for NDB API credentials and ensure the secret exists there (create it or copy from the old namespace):

```bash
kubectl create namespace ndb-credentials
# If the secret was in another namespace (e.g. default), copy it:
kubectl get secret <your-ndb-secret-name> -n <old-namespace> -o yaml | \
  sed 's/namespace: .*/namespace: ndb-credentials/' | kubectl apply -f -
# Or create a new secret in ndb-credentials with your NDB API credentials.
```

Use a single secret name (e.g. `ndb-api-secret`) that you will reference in the new NDBServer.

### Step 3: Remove existing namespaced NDBServer(s)

You must delete all NDBServer resources before upgrading, or the CRD update (namespaced → cluster-scoped) will be rejected.

```bash
kubectl delete ndbserver --all -A
```

**Database** resources can stay; they will reconcile again once the new cluster-scoped NDBServer exists and `ndbRef` matches its name.

### Step 4: Upgrade the operator to 0.5.3

Install or deploy the ndb-operator **0.5.3** release (manifests, Helm, or OLM). This installs the updated CRD with `scope: Cluster` and the new operator image.

### Step 5: Create cluster-scoped NDBServer(s)

Create one NDBServer per logical NDB server. Use the **same `metadata.name`** as before if you want existing **Database** resources (which reference NDBServer by name in `spec.ndbRef`) to work without changes.

**NOTE ⚠️ : when the NDB server yaml file is applied and it shows error that CRD spec.CredentialSecret field is incorrect then please run below command to delete existing NDB server CRD and install new one from the installation commands.**
```bash
kubectl delete crds ndbservers.ndb.nutanix.com
```
Example:

```yaml
apiVersion: ndb.nutanix.com/v1alpha1
kind: NDBServer
metadata:
  name: ndb   # same name as before so existing Databases need no change
spec:
  server: "https://<ndb-server>:8443/era/v0.9"
  credentialSecretRef:
    name: ndb-api-secret
    namespace: ndb-credentials
```

Apply with `kubectl apply -f <file>`. No `namespace` in metadata (cluster-scoped).

### Step 6: Verify Databases

Ensure each **Database** has `spec.ndbRef` set to the cluster-scoped NDBServer’s **name** (e.g. `ndb`). If you used the same NDBServer name as before, no change is needed. Check reconciliation:

```bash
kubectl get databases -A
kubectl get ndbserver
```

### Summary

| Step | Action |
|------|--------|
| 1 | Back up NDBServer(s) (`kubectl get ndbserver -A -o yaml`) |
| 2 | Create `ndb-credentials` namespace and put/copy NDB API secret there |
| 3 | Delete all namespaced NDBServer(s) (`kubectl delete ndbserver --all -A`) |
| 4 | Install ndb-operator **0.5.3** |
| 5 | Create new cluster-scoped NDBServer(s) with `credentialSecretRef` and same name(s) as before |
| 6 | Confirm Databases reconcile (same `ndbRef` if you kept the same NDBServer name) |

---

## Upgrading from namespaced NDBServer (earlier versions)

If you are on an older version where **NDBServer** was namespaced, the same conceptual migration applies: NDBServer is now cluster-scoped and uses `credentialSecretRef` (name + namespace) instead of `credentialSecret`. Follow the steps in [Upgrading from 0.5.2 to 0.5.3](#upgrading-from-052-to-053-ndbserver-namespaced--cluster-scoped) above.

---

## Developement

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make generate manifests
```
Add the CRDs to the Kubernetes cluster
```sh
make install
```
Run your controller locally (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTES:** 
1. You can also run this in one step by running: `make install run`
2. Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

### Building and pushing to an image registry
Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/ndb-operator:tag
```

### Deploy the operator pushed to an image registry
Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/ndb-operator:tag
```
---
## Uninstallation / Cleanup
Uninstall the operator based on the installation/deployment  environment

### Running outside the cluster
```sh
# Stops the controller process
ctrl + c
# Uninstalls the CRDs
make uninstall
```

### Running inside the cluster
```sh
# Removes the deployment, crds, services and rbac entities
make undeploy
```

### Running using Helm charts
```sh
# NAME: name of the release created during installation
helm uninstall NAME
```

### Running on Openshift
```sh
operator-sdk cleanup ndb-operator --delete-all
```

---

## How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/). It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/)
which provides a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

A custom resource of the kind Database is created by the reconciler, followed by a Service and an Endpoint that maps to the IP address of the database instance provisioned. Application pods/deployments can use this service to interact with the databases provisioned on NDB through the native Kubernetes service.

Pods can specify an initContainer to wait for the service (and hence the database instance) to get created before they start up.
```yaml
  initContainers:
  - name: init-db
    image: busybox:1.28
    command: ['sh', '-c', "until nslookup <<Database CR Name>>-svc.$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local; do echo waiting for database service; sleep 2; done"]
```
---

## Contributing
See the [contributing docs](CONTRIBUTING.md)

---

## Support
This code is developed in the open with input from the community through issues and PRs. A Nutanix engineering team serves as the maintainer. Documentation is available in the project repository. Issues and enhancement requests can be submitted in the [Issues tab of this repository](../../issues). Please search for and review the existing open issues before submitting a new issue.

---

## License
Copyright 2022-2023 Nutanix, Inc.

The project is released under version 2.0 of the [Apache license](http://www.apache.org/licenses/LICENSE-2.0).
