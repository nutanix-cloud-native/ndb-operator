# Nutanix Database Service Operator for Kubernetes
The NDB operator brings automated and simplified database administration, provisioning, and life-cycle management to Kubernetes.

## Getting Started
### Pre-requisites
1. Era / NDB on-prem [installation](https://portal.nutanix.com/page/documents/details?targetId=Nutanix-Era-User-Guide-v2_4:top-era-installation-c.html).
2. A Kubernetes cluster to run against, which should have network connectivity to the NDB installation. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** The operator will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).
3. The operator-sdk installed.
4. A clone of the source code (this repository).
### Installation and Running on the cluster
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`      
<br>
### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make generate manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

<br>

### Using the Operator

1. To create instances of custom resources (provision databases), edit [ndb_v1alpha1_database.yaml](config/samples/ndb_v1alpha1_database.yaml) file with the NDB installation and database instance details and run:

```sh
kubectl apply -f config/samples/ndb_v1alpha1_database.yaml
```
2. To delete instances of custom resources (deprovision databases) run:

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
  ndb:
    # Cluster id of the cluster where the Database has to be provisioned
    # Can be fetched from the GET /clusters endpoint
    clusterId: "Nutanix Cluster Id" 
    # Credentials for NDB installation
    credentials:
      loginUser: admin
      password: "NDB Password"
      sshPublicKey: "SSH Key"
    # The NDB Server
    server: https://[NDB IP]:8443/era/v0.9

  databaseInstance:
    # The database instance name on NDB
    databaseInstanceName: "Database Instance Name"
    # Names of the databases on that instance
    databaseNames:
      - database_one
      - database_two
      - database_three
    # Password for the database
    password: qwertyuiop
    size: 10
    timezone: "UTC"
    type: postgres
```

<br>

### Building and pushing to an image registry  
Build and push your image to the location specified by `IMG`:
	
```sh
make docker-build docker-push IMG=<some-registry>/ndb-operator:tag
```

<br>

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
UnDeploy the controller to the cluster:

```sh
make undeploy
```

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster 

## Contributing
See the [contributing docs](CONTRIBUTING.md).

## Support
### Community Plus

This code is developed in the open with input from the community through issues and PRs. A Nutanix engineering team serves as the maintainer. Documentation is available in the project repository.

Issues and enhancement requests can be submitted in the [Issues tab of this repository](../../issues). Please search for and review the existing open issues before submitting a new issue.

## License

Copyright 2021-2022 Nutanix, Inc.

The project is released under version 2.0 of the [Apache license](http://www.apache.org/licenses/LICENSE-2.0).
