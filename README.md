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

1. Create file "secrets.yaml" to store the secrets that can be used by the custom resources:

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
kubectl apply -f <path/to/secrets.yaml>
```
You can optionally verify that they have been created:

```sh
kubectl get secrets
```

2. To create a Database CR, update these fields in the "spec" section of [ndb_v1alpha1_database.yaml](config/samples/ndb_v1alpha1_database.yaml)
    <br /> a. "server"               : NDB Server IP
    <br /> b. "clusterId"            : Nutanix Cluster Id
    <br /> c. "databaseInstanceName" : Database Instance Name

3. Finally, run this command to provision the database using NDB Operator:
```sh
kubectl apply -f config/samples/ndb_v1alpha1_database.yaml
```
4. To delete the Database CR (deprovision database) run:

```sh
kubectl delete -f config/samples/ndb_v1alpha1_database.yaml
```
The [ndb_v1alpha1_database.yaml](config/samples/ndb_v1alpha1_database.yaml) is described as follows:
```yaml
apiVersion: ndb.nutanix.com/v1alpha1
kind: Database
metadata:
  # This name that will be used within the kubernetes cluster
  name: db
spec:
  # NDB server specific details
  ndb:
    # Cluster id of the cluster where the Database has to be provisioned
    # Can be fetched from the GET /clusters endpoint
    clusterId: "Nutanix Cluster Id"
    # Credentials secret name for NDB installation
    # data: username, password,
    # stringData: ca_certificate
    credentialSecret: ndb-secret-name
    # The NDB Server
    server: https://[NDB IP]:8443/era/v0.9
    # Set to true to skip SSL verification, default: false.
    skipCertificateVerification: true
  # Database instance specific details (that is to be provisioned)
  databaseInstance:
    # The database instance name on NDB
    databaseInstanceName: "Database-Instance-Name"
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
