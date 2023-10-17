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
## Getting Started
### Pre-requisites
1. NDB [installation](https://portal.nutanix.com/page/documents/details?targetId=Nutanix-NDB-User-Guide-v2_5:Nutanix-NDB-User-Guide-v2_5).
2. A Kubernetes cluster to run against, which should have network connectivity to the NDB installation. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** The operator will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).
3. The operator-sdk installed.
4. A clone of the source code ([this](https://github.com/nutanix-cloud-native/ndb-operator) repository).
5. Installing the cert-manager. Please follow the instructions [here](https://cert-manager.io/docs/installation/)

### Installation and Running on the cluster
Deploy the controller on the cluster:

```sh
make deploy
```

### Using the Operator

1. Create the secrets that will be used by the NDBServer and Database resources:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: ndb-secret-name
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
type: Opaque
stringData:
  password: password-for-the-database-instance
  ssh_public_key: SSH-PUBLIC-KEY

```

Apply the secrets:

```
kubectl apply -f <path/to/secrets-manifest.yaml>
```
You can optionally verify that they have been created:

```sh
kubectl get secrets
```

2. To create a NDBServer resource manifest that holds the information about the NDB setup, the fields in the `spec` section of the sample manifest [ndb_v1alpha1_ndbserver.yaml](config/samples/ndb_v1alpha1_ndbserver.yaml) should be updated. The file is described as follows:

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
spec:
    # Name of the secret that holds the credentials for NDB: username, password and ca_certificate (created in step 1) 
    credentialSecret: ndb-secret-name
    # NDB Server's API URL
    server: https://[NDB IP]:8443/era/v0.9
    # Set to true to skip SSL certificate validation, should be false if ca_certificate is provided in the credential secret.
    skipCertificateVerification: true

```

3. Run this command to create the NDBServer resource:
```sh
kubectl apply -f config/samples/ndb_v1alpha1_ndbserver.yaml
```

4. To create a Database resource manifest that holds the information about the Database, the fields in the `spec` section of the sample manifest [ndb_v1alpha1_database.yaml](config/samples/ndb_v1alpha1_database.yaml) should be updated. The file is described as follows:

```yaml
apiVersion: ndb.nutanix.com/v1alpha1
kind: Database
metadata:
  # This name that will be used within the kubernetes cluster
  name: db
spec:
  # Name of the NDBServer resource created in step 3
  ndbRef: ndb
  # Database instance specific details (that is to be provisioned)
  databaseInstance:
    # Cluster id of the cluster where the Database has to be provisioned
    # Can be fetched from the GET /clusters endpoint
    clusterId: "Nutanix Cluster Id"
    # The database instance name on NDB
    Name: "Database-Instance-Name"
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
    additionalArguments:                # Optional black, can specify additional arguments that are unique to database engines.
      listener_port: 8080

```

5. Run this command to create the Database resource:

```sh
kubectl apply -f config/samples/ndb_v1alpha1_database.yaml
```
6. To delete the Database resource (deprovision database) run:

```sh
kubectl delete -f config/samples/ndb_v1alpha1_database.yaml
```
7. To delete the NDBServer resource run:

```sh
kubectl delete -f config/samples/ndb_v1alpha1_ndbserver.yaml
```
Below are the various optional addtionalArguments you can specify along with examples of their corresponding values. Arguments that have defaults will be indicated.

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

## Developement

### Installation and Running the controller locally
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller locally (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make generate manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

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

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
To remove the controller from the cluster:

```sh
make undeploy
```

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
See the [contributing docs](CONTRIBUTING.md).

## Support
### Community Plus

This code is developed in the open with input from the community through issues and PRs. A Nutanix engineering team serves as the maintainer. Documentation is available in the project repository.

Issues and enhancement requests can be submitted in the [Issues tab of this repository](../../issues). Please search for and review the existing open issues before submitting a new issue.

## License

Copyright 2022-2023 Nutanix, Inc.

The project is released under version 2.0 of the [Apache license](http://www.apache.org/licenses/LICENSE-2.0).
