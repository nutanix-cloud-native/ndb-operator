<h2>Background</h2>
Kubernetes
An open-source container orchestration technology called Kubernetes is used to automatically deploy, scale, and manage containerized applications. Developers can use Kubernetes to distribute and control containerized applications across a dispersed network of servers or PCs. To ensure that the actual state of an application matches the desired state, it uses a declarative model to express the desired state and automatically manages the containerized components. Kubernetes can be operated on public or private cloud infrastructure as well as in-house data centers and offers a wide range of functionality for managing containerized applications, such as autonomous scaling, rolling updates, self-healing, service discovery, and load balancing.

<h3>Nutanix Database Service</h3>

A hybrid multi-cloud database-as-a-service for Microsoft SQL Server, Oracle Database, PostgreSQL, MongoDB, and MySQL, among other databases, is called Nutanix Database Service. It allows for the efficient management of hundreds to thousands of databases, the quick creation of new ones, and the automation of time-consuming administration activities like patching and backups. Users can also choose certain operating systems, database versions, and extensions to satisfy application and compliance requirements. Customers from all around the world have optimized their databases across numerous locations and sped up software development using Nutanix Database Service.

<h3>Features offered by NDB Service:</h3>

![img1](https://user-images.githubusercontent.com/96166947/230685233-d4eb8056-730b-4cec-b269-001c62a1629c.png)

<ol>
<li>Nutanix NDB is a distributed NoSQL database service that is part of the Nutanix platform. Some of the key features of NDB include highly scalable architecture, distributed data storage, support for multiple data models, consistent data, fast data access, automatic sharding, real-time analytics, high availability and fault tolerance, and strong security features.</li>

<li>With its ability to scale up or down the number of nodes in a cluster, Nutanix NDB provides highly scalable architecture without any downtime. Its distributed architecture ensures high availability and fault tolerance, while its support for multiple data models makes it a versatile database service for a wide range of use cases. Additionally, NDB supports strong consistency and fast data access by caching frequently accessed data in memory, which helps reduce the number of disk reads and improves query performance.</li>

<li>NDB also provides automatic sharding, which helps ensure that your database can handle large amounts of data. You can use graph queries to analyze relationships between data in real-time, which can help you make more informed decisions. Furthermore, NDB offers high availability and fault tolerance through its distributed architecture and replication features. Lastly, NDB provides strong security features, including role-based access control, data encryption at rest, and network security features.</li>

</ol>

![img2](https://user-images.githubusercontent.com/96166947/230685113-00e22821-378b-44a9-ad74-864812275014.jpg)

<h3>NDB Kubernetes Operator</h3>

The NDB Kubernetes Operator is an innovative tool created by Nutanix to streamline the management and operation of the Nutanix NDB (NoSQL database) on Kubernetes clusters.

With the NDB Kubernetes Operator, deploying and managing NDB clusters on Kubernetes has never been easier, as it eliminates the need to manually configure and manage the underlying infrastructure. Built on the Kubernetes operator framework, it offers a declarative way to manage the lifecycle of NDB clusters and other related resources.

One of the key benefits of the operator is that it simplifies the management of NDB clusters by automating common tasks, such as cluster creation, scaling, upgrading, backup, and recovery. It also offers a high degree of flexibility and customization, allowing you to configure various aspects of the cluster, such as storage, networking, and security.

Another advantage of the NDB Kubernetes Operator is its seamless integration with other Kubernetes tools and resources, such as Helm charts, Kubernetes secrets, and Kubernetes ConfigMaps. This integration makes it easy to integrate NDB into your existing Kubernetes-based infrastructure and workflows, providing a hassle-free solution for managing your database clusters.

Overall, the NDB Kubernetes Operator is a powerful and flexible tool for managing NDB clusters on Kubernetes, freeing you up to focus on your application logic rather than infrastructure management. Its automation capabilities and integration with other Kubernetes tools make it a must-have tool for developers and administrators looking to simplify and streamline their database management on Kubernetes.

<h2>Existing Architecture and Problem Statement</h2>
<h3>Problem Statement: Refactor models to keep profiles (software, compute, network, etc) as optional and use default if not specified</h3>

The NDB Kubernetes operator currently uses default compute, network and OS software profiles while provisioning the database. Refactor this module to include optional fields and only if absent, fall back to default.

<h3>NDB Architecture</h3>

![img3](https://user-images.githubusercontent.com/96166947/230685136-f643fff0-a55e-4186-b551-371f2536e677.png)


Microsoft SQL Server, Oracle Database, PostgreSQL, MySQL, and MongoDB are just a few of the databases that can have high availability, scalability, and speed thanks to the distributed architecture of the Nutanix Database Service. The hyper-converged infrastructure from Nutanix, which offers a scalable and adaptable platform for handling enterprise workloads, is the foundation around which the architecture is built.

There are various layers in the architecture of the Nutanix Database Service. The Nutanix hyperconverged infrastructure is the basic layer that provides the storage, computing, and networking resources needed to run the databases. The Nutanix Acropolis operating system, which offers the essential virtualization and administration features, sits on top of this layer.

The Nutanix Era layer, which is located above the Nutanix Acropolis layer, offers the Nutanix Database Service the ability to manage databases throughout their existence. The Nutanix Era Manager, a centralized management console that offers a single point of access for controlling the databases across several clouds and data centers, is included in this tier.

The Nutanix Era Orchestrator, which is in charge of automating the provisioning, scaling, patching, and backup of the databases, is another component of the Nutanix Era layer. The Orchestrator offers a declarative approach for specifying the desired state of the databases and is built to work with a variety of databases.

The Nutanix Era Application, a web-based interface that enables database administrators and developers to quickly provision and administer the databases, is the final component of the top layer. A self-service interface for installing databases as well as a number of tools for tracking and troubleshooting database performance are offered by the Era Application.

<h2>Design & Workflow</h2>
Large amounts of data may be handled by the highly scalable, fault-tolerant, and consistent Nutanix NDB NoSQL database. It is a distributed database created to be installed over several cluster nodes. A portion of the data is stored on each node in the cluster, and the data is replicated across several nodes to guarantee high availability.

Configure your Nutanix cluster: We need to configure your Nutanix cluster to support NDB. This includes setting up the storage and network configurations, configuring the NDB nodes, and defining the replication factor.

Create a table: We need to create a table in NDB to store your data. This includes defining the schema, specifying the replication factor, and configuring any other options you need.

Write your code: We need to write your code to interact with the NDB cluster. This includes inserting and retrieving data, as well as performing more complex operations such as querying, indexing, and data aggregation.

Test your code: We need to test your code to ensure that it works as expected. This includes testing basic operations such as creating and retrieving data, as well as testing more complex operations such as queries and data aggregation.

Monitor your cluster: We need to monitor your NDB cluster to ensure that it is performing as expected. This includes monitoring resource usage, handling errors and exceptions, and optimizing performance.

Optimize your cluster: We need to optimize your NDB cluster over time to ensure that it continues to meet your needs. This includes tuning the configuration, optimizing queries, and scaling the cluster as needed.

Backup and recovery: We need to establish backup and recovery procedures to ensure that your data is protected against data loss or corruption. This includes regularly backing up your data, testing your backups, and establishing procedures for recovering data in case of a disaster.

<img width="600" alt="img4" src="https://user-images.githubusercontent.com/96166947/230685149-31800d7a-c3cd-4879-ad29-086ac2648cf4.png">


<h2>Potential Design Patterns, Principles, and Code Refactoring strategies</h2>


Employing Clean Code Practices + Design Patterns:

DRY (Do Not Repeat Yourself):
The initial draft of using 4 if & else statements for each of the profiles (namely: software, compute, network, and storage) and performing the same checks and code again proved to be in direct opposition to the DRY approach whose repetitiveness can be seen from the below snippet:
```
            //Custom Software Profile Check and overriding the default values
            softwareProfile := dbSpec.Instance.Profiles.Software
            if softwareProfile == (Profile{}) {
                log.Info("No enrichment for software profiles as no custom profile received for Software. Hence, proceeding with default OOB software profile")
            } else {
                isValidProfile, matchedProfile, errorThroughChecks := performProfileAvailabilityCheck(ctx, dbEngineSpecificProfiles, softwareProfile, PROFILE_TYPE_SOFTWARE)
                if errorThroughChecks != nil {
                    //log.Error(err, "")
                    return errorThroughChecks
                }
                if isValidProfile {
                    profilesMap[PROFILE_TYPE_SOFTWARE] = matchedProfile
                }
            }

            //Custom Compute Profile Check and overriding the default values
            computeProfile := dbSpec.Instance.Profiles.Compute
            if computeProfile == (Profile{}) {
                log.Info("No enrichment for compute profiles as no custom profile received for Compute. Hence, proceeding with default OOB compute profile")
            } else {
                isValidProfile, matchedProfile, errorThroughChecks := performProfileAvailabilityCheck(ctx, genericProfiles, computeProfile, PROFILE_TYPE_COMPUTE)
                if errorThroughChecks != nil {
                    //log.Error(err, "")
                    return errorThroughChecks
                }
                if isValidProfile {
                    profilesMap[PROFILE_TYPE_COMPUTE] = matchedProfile
                }
            }

            //Custom Network Profile Check and overriding the default values
            networkProfile := dbSpec.Instance.Profiles.Network
            if networkProfile == (Profile{}) {
                log.Info("No enrichment for network profiles as no custom profile received for Network. Hence, proceeding with default OOB network profile")
            } else {
                isValidProfile, matchedProfile, errorThroughChecks := performProfileAvailabilityCheck(ctx, dbEngineSpecificProfiles, networkProfile, PROFILE_TYPE_NETWORK)
                if errorThroughChecks != nil {
                    //log.Error(err, "")
                    return errorThroughChecks
                }
                if isValidProfile {
                    profilesMap[PROFILE_TYPE_NETWORK] = matchedProfile
                }
            }

            //Custom DbParam Profile Check and overriding the default values
            dbParamProfile := dbSpec.Instance.Profiles.DbParam
            if dbParamProfile == (Profile{}) {
                log.Info("No enrichment for database parameter profiles as no custom profile received for it. Hence, proceeding with default OOB dbParam profile")
            } else {
                isValidProfile, matchedProfile, errorThroughChecks := performProfileAvailabilityCheck(ctx, dbEngineSpecificProfiles, dbParamProfile, PROFILE_TYPE_DATABASE_PARAMETER)
                if err != nil {
                    //log.Error(err, "")
                    return errorThroughChecks
                }
                if isValidProfile {
                    profilesMap[PROFILE_TYPE_DATABASE_PARAMETER] = matchedProfile
                }
            }
        }
        return
    }
```
Hence, we create an array of profiles, and iterating over those profiles and calling a clean modular function by identifying the common key points from the above snippet helps in eliminating code duplication and makes the code more, readable, and open for extension by just adding new profile name to the list of profiles and easily maintainable! This approach can be viewed in EnrichProfilesMap() and delegate functions.

Refactoring + Delegation using Facade Design Pattern:

By default, for populating profiles, GetOOBProfiles() was called which did the task of fetching all profiles and populating default profile values. However, with the advent of custom profiles being provided from YAML, we suggest refactoring GetOOBProfiles() to EnrichProfilesMap() which will perform the same task as GetOOBProfiles did but in addition override the default values to input profiles provided after performing checks such as:
(1) Emptiness / Null checks for profiles
(2) Performing matching of the Id/VersionId for the custom profile provided & failing the database provisioning request in case of a match is not found.
Thus, for each of the checkers, we create modular functions and delegate the task of performing the above-mentioned checks.

The flow proceeds as:

EnrichProfilesMap() 
Performs the task to set default values and override the default values by calling the below functions.  

PerformProfileMatchingAndEnrichProfiles() uses
[isEmptyProfile() + GetAppropriateProfileForType() + GetTopologyForProfileType() in filtering profiles]
Once, the user input is received, the input is delegated for checks and performing matching if the profile exists. Additionally, other factory methods are also used for performing the matching of profiles

EnrichProfileMapForProfileType()
Performs the final overriding of the default profile with the matched profile. Additionally, it cancels the database provisioning request for unmatched profiles.

Moreover, in Tests, new tests have been added which indicate their functionality through their names adhering to the Clean Coding Naming principle:
TestEnrichAndGetProfilesWhenCustomProfilesMatch()
TestEnrichAndGetProfilesWhenInvalidCustomProfilesProvided()

Additionally, we can also view the creation of specific functions as the alignment with the Facade Design Pattern that delegates the task of performing specific actions to respective functions. 

Alternatively, a closer look at the initial draft and performing Cyclomatic Complexity checks indicate that our approach has breached the threshold value of permissible complexity. Hence, to tackle this problem, we will change the filtering profile functions to perform matching based on Id and further based on versionId and eliminating factory methods like getTopology(), getProfileForType(), and a few checker functions that will aid in reducing the cyclomatic complexity automatically.


<h2> Modifications </h2>

<h3> \ndb-operator\api\v1alpha1\ndb_api_helpers.go </h3>

<h4>Functions Changed</h4>
<ol>
<li><h4> GenerateProvisioningRequest </h4>

<b>Previous Working :</b> This function generates and returns a request for provisioning a database (and a dbserver vm) on NDB and uses default compute, software, network, databaseParams profiles

<br><b>Enhanced Working :</b> This function generates and returns a request for provisioning a database (and a dbserver vm) on NDB and if user has provided custom profiles in "\ndb-operator\config\samples\ndb_v1alpha1_database.yaml", it will use those profiles to create the provisioning request or it will fall back to default profiles

<br><b>Previous Code :</b>
```
// Fetch the OOB profiles for the database
profilesMap, err := GetOOBProfiles(ctx, ndbclient, dbSpec.Instance.Type)
if err != nil {
    log.Error(err, "Error occurred while getting OOB profiles", "database name", dbSpec.Instance.DatabaseInstanceName, "database type", dbSpec.Instance.Type)
    return
}
```

<br><b>New Code :</b>
```
// Fetch upto date profiles for the database
	profilesMap, err := MatchAndGetProfiles(ctx, ndbclient, dbSpec.Instance.Type, dbSpec.Instance.Profiles)
	if err != nil {
		log.Error(err, "Error occurred while enriching and getting profiles", "database name", dbSpec.Instance.DatabaseInstanceName, "database type", dbSpec.Instance.Type)
		return
	}
```

<br><b>Explanation for the change :</b>
Replaced the call for GetOOBProfiles function with MatchAndGetProfiles due to added functionality of populating the profileMap with all the values for the profileType available
</li>

<li><h4> MatchAndGetProfiles </h4>

<b>Previous Working :</b> This function was not there previously.

<br><b>Enhanced Working :</b> now this function first fetches all the profiles from NDB API. Then for each profile type it checks if a profile is provided in the YAML file or not. If it is provided in the YAML file, then this function calls GetProfileByType and matchProfiles functions to do the further work. If it is not provided in the YAML file, then this function calls PopulateDefaultProfiles function to assign default profiles.

<br><b>Previous Code :</b> N/A

<br><b>New Code :</b>
```
  func MatchAndGetProfiles(ctx context.Context, ndbclient *ndbclient.NDBClient, dbType string, profiles Profiles) (profileMap map[string]ProfileResponse, err error) {

	log := ctrllog.FromContext(ctx)

	// Map of profile type to profiles
	profileMap = make(map[string]ProfileResponse)

	allProfiles, err := GetAllProfiles(ctx, ndbclient)

	if err != nil {
		return
	}

	log.Info("Received Input Profiles = ", "Received Input Profiles", profiles)
	profileOptions := [...]string{PROFILE_TYPE_COMPUTE, PROFILE_TYPE_SOFTWARE, PROFILE_TYPE_NETWORK, PROFILE_TYPE_DATABASE_PARAMETER, PROFILE_TYPE_STORAGE}
	for _, profileType := range profileOptions {
		if profiles == (Profiles{}) {
			err = PopulateDefaultProfile(ctx, profileMap, profileType, allProfiles, dbType)
		} else {
			profile := GetProfileByType(profileType, profiles)
			err = matchProfiles(ctx, profileType, profile, allProfiles, profileMap, dbType)
		}
		if err != nil {
			return
		}
	}

	return

}
```

<br><b>Explanation for the change :</b> We need a function that checks if profiles are there in YAML file for or not and depending on that we need to make the further decisions. This function helps us to check that and delegate the next tasks.
</li>
<li><h4> matchProfiles </h4>

<b>Previous Working :</b> This function was not there previously.

<br><b>Enhanced Working :</b> This function does:
	(a) Id level matching with profiles in allProfiles
	(b) If Id level match is successful, flow proceeds to match based on versionId
		When matched, the latestVersionId is overridden with the versionId as it is this attribute while dbProvisioning which is used for
		profileType versionId.

<br><b>Previous Code :</b> N/A

<br><b>New Code :</b>
```
func matchProfiles(ctx context.Context, profileType string, profile Profile, allProfiles []ProfileResponse, profilesMap map[string]ProfileResponse, dbType string) (err error) {
	log := ctrllog.FromContext(ctx)
	var idMatchedProfile []ProfileResponse
	var matchedVersion []Version
	if isEmptyProfile(profile) {
		err = PopulateDefaultProfile(ctx, profilesMap, profileType, allProfiles, dbType)
		return
	}
	log.Info("Performing profile matching for profileType => ", "profileType", profileType)
	// match based on ID
	idMatchedProfile = util.Filter(allProfiles, func(p ProfileResponse) bool { return p.Id == profile.Id && p.Type == profileType })
	// matching based on versionID
	if len(idMatchedProfile) > 0 {
		matchedVersion = util.Filter(idMatchedProfile[0].Versions, func(versions Version) bool { return versions.Id == profile.VersionId })
		// when versionID level match found, override latestVersionId as it is used in the database provisioning request
		if len(matchedVersion) > 0 {
			log.Info("Id and VersionId matched for profileType", "profileType", profileType)
			idMatchedProfile[0].LatestVersionId = profile.VersionId
		}
	}
	err = PopulateProfileOfType(ctx, profilesMap, profileType, allProfiles, dbType, idMatchedProfile)
	return
}
```

<br><b>Explanation for the change :</b> We use profiles in YAML file only when they have correct Id. To perform this essential task of checking the correctness of Id and after that populating the versionId, we have created this function. After checking the correctness, this function delegates the further tasks to PopulateProfileOfType function.
</li>
<li><h4> PopulateProfileOfType </h4>

<b>Previous Working :</b> This function was not there previously.

<br><b>Enhanced Working :</b> This function performs the task of populating profileMap with response (matching result) received for the profileType.

<br><b>Previous Code :</b> N/A

<br><b>New Code :</b>
```
func PopulateProfileOfType(ctx context.Context, profileMap map[string]ProfileResponse, profileType string, allProfiles []ProfileResponse, dbType string, response []ProfileResponse) (err error) {
	log := ctrllog.FromContext(ctx)
	// if response is empty, it indicates no matching profile found; hence set the default OOB profile for that type
	if len(response) == 0 {
		err = fmt.Errorf("No matching profile found for profileType = %s", profileType)
		log.Info("Error Occurred. No enrichment performed for profile = ", "profileType", profileType)
		return
	}
	log.Info("Going to populate profile value in profilesMap for profileType = ", "profileType", profileType)
	profileMap[profileType] = response[0]
	return
}

```

<br><b>Explanation for the change :</b> Since this function is called only after the sanity check of Id and versionId, this function just has to populate the profiles by their respective types.
</li>
<li><h4> PopulateDefaultProfile </h4>

<b>Previous Working :</b> This function was not there previously.

<br><b>Enhanced Working :</b> This method populates profileMap with the default value for the profileType.

<br><b>Previous Code :</b> N/A

<br><b>New Code :</b>
```
func PopulateDefaultProfile(ctx context.Context, profileMap map[string]ProfileResponse, profileType string, allProfiles []ProfileResponse, dbType string) (err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Going to set default profile value for profileType = ", "profileType", profileType)
	genericProfiles := util.Filter(allProfiles, func(p ProfileResponse) bool { return p.EngineType == DATABASE_ENGINE_TYPE_GENERIC })
	dbEngineSpecificProfiles := util.Filter(allProfiles, func(p ProfileResponse) bool { return p.EngineType == GetDatabaseEngineName(dbType) })
	response, err := GetDefaultProfileForType(genericProfiles, dbEngineSpecificProfiles, profileType)
	if err != nil {
		return
	}
	profileMap[profileType] = response[0]
	return
}
```

<br><b>Explanation for the change :</b> Since this function is called only if profiles are missing in the YAML file or have wrong profile Id, this function just has to populate the default out of the box profiles for each profile type. This function calls GetDefaultProfileForType function to get profiles for each profile type.
</li>
<li><h4> GetDefaultProfileForType </h4>

<b>Previous Working :</b> This function was not there previously.

<br><b>Enhanced Working :</b> This function gets default profile for each profile type from the result of GET API call to the NDB API.

<b>Previous Code :</b> N/A

<br><b>New Code :</b>
```
func GetDefaultProfileForType(genericProfiles []ProfileResponse, dbEngineSpecificProfiles []ProfileResponse, profileType string) (profile []ProfileResponse, err error) {
	switch profileType {
	case PROFILE_TYPE_COMPUTE:
		profile = util.Filter(genericProfiles, func(p ProfileResponse) bool {
			return p.Type == PROFILE_TYPE_COMPUTE && strings.Contains(strings.ToLower(p.Name), "small")
		})
		break
	case PROFILE_TYPE_SOFTWARE:
		profile = util.Filter(dbEngineSpecificProfiles, func(p ProfileResponse) bool { return p.Type == PROFILE_TYPE_SOFTWARE && p.Topology == TOPOLOGY_SINGLE })
		break
	case PROFILE_TYPE_NETWORK:
		profile = util.Filter(dbEngineSpecificProfiles, func(p ProfileResponse) bool { return p.Type == PROFILE_TYPE_NETWORK })
		break
	case PROFILE_TYPE_DATABASE_PARAMETER:
		profile = util.Filter(dbEngineSpecificProfiles, func(p ProfileResponse) bool { return p.Type == PROFILE_TYPE_DATABASE_PARAMETER })
		break
	case PROFILE_TYPE_STORAGE:
		profile = util.Filter(genericProfiles, func(p ProfileResponse) bool { return p.Type == PROFILE_TYPE_STORAGE })
		break
	default:
		return
	}
	if len(profile) == 0 {
		err = errors.New("oob profile: one or more OOB profile(s) were not found")
		return
	}
	return
}
```

<br><b>Explanation for the change :</b> We have to find appropriate default profile from the API call result. With the help of some util functions, this function performs this filtering task.
</li>
<li><h4> GetProfileByType </h4>

<b>Previous Working :</b> This function was not there previously.

<br><b>Enhanced Working :</b> This function just returns the appropriate costant value from the pre-defined object structures.

<b>Previous Code :</b> N/A

<br><b>New Code :</b>
```
func GetProfileByType(profileType string, profiles Profiles) Profile {
	defaultEmptyProfile := Profile{}
	switch profileType {
	case PROFILE_TYPE_COMPUTE:
		return profiles.Compute
	case PROFILE_TYPE_SOFTWARE:
		return profiles.Software
	case PROFILE_TYPE_NETWORK:
		return profiles.Network
	case PROFILE_TYPE_DATABASE_PARAMETER:
		return profiles.DbParam
	default:
		return defaultEmptyProfile
	}
}
```

<br><b>Explanation for the change :</b> We use pre-defined object structures for reference everywhere. This function helps us to make that references correctly.
</li>
</ol>
<h2> Test Plan </h2>

<h3> Test Case Scenario 1 </h3>

<br><b>Test case name :</b> Provisioning of appropriate database based on provided software/compute/network/dbParams profiles

<br><b>Description :</b> This test case verifies that the appropriate database is provisioned based on the provided software/compute/network/dbParams profiles as input through YAML file, as expected.

<br><b>Pre-conditions :</b>
<ul>
<li>Pre-requisites are installed</li>
<li>Docker Desktop Application is running</li>
<li>Kubernetes cluster is up</li>
<li>Nutanix Test Drive is active and the cluster id and other credentials are present inside "\ndb-operator\config\samples\ndb_v1alpha1_database.yaml" and "\ndb-operator\config\samples\secret.yaml"</li>
<li>The software/compute/network/dbParams profiles are available for input in a profiles section inside "\ndb-operator\config\samples\ndb_v1alpha1_database.yaml"</li>
</ul>

<br><b>Test steps :</b>
<ul>
<li>Run command "make install run" in the root directory of the project
<li>Create secrets with command "kubectl apply -f .\config\samples\secret.yaml"
<li>Provision the database with command "kubectl apply -f .\config\samples\ndb_v1alpha1_database.yaml"
<li>Check if the appropriate database has been provisioned on the Nutanix test drive
<li>Verify that the compute/software/network/dbParams profiles of the database match the expected values based on the input parameters
</ul>

<br><b>Expected results :</b>
<ul>
<li>The system provisions the appropriate database based on the configurations specified in "\ndb-operator\config\samples\ndb_v1alpha1_database.yaml"
<li>The the compute/software/network/dbParams profiles match the expected values based on the input parameters
<li>The test case passes successfully
</ul>


<h3> Test Case Scenario 2 </h3>

<br><b>Test case name :</b> Throwing error if invalid software/compute/network/dbParams profiles are given as input

<br><b>Description :</b> This test case verifies that error is thrown if invalid software/compute/network/dbParams profiles are provided as input through YAML file.

<br><b>Pre-conditions :</b>
<ul>
<li>Pre-requisites are installed
<li>Docker Desktop Application is running
<li>Kubernetes cluster is up
<li>Nutanix Test Drive is active and the cluster id and other credentials are present inside "\ndb-operator\config\samples\ndb_v1alpha1_database.yaml" and "\ndb-operator\config\samples\secret.yaml"
<li>The software/compute/network/dbParams profiles are available for input in a profiles section inside "\ndb-operator\config\samples\ndb_v1alpha1_database.yaml"
</ul>

<br><b>Test steps :</b>
<ul>
<li>Run command "make install run" in the root directory of the project
<li>Create secrets with command "kubectl apply -f .\config\samples\secret.yaml"
<li>Provision the database with command "kubectl apply -f .\config\samples\ndb_v1alpha1_database.yaml"
<li>Check if the database has not been provisioned on the Nutanix test drive
<li>Verify that the error is thrown on the command prompt
</ul>

<br><b>Expected results :</b>
<ul>
<li>The system does not provision the database
<li>The error is thrown saying that the id/version id of software/compute/network/dbParams profiles is invalid
<li>The test case passes successfully
</ul>

<h3> Test Case Scenario 3 </h3>

<br><b>Test case name :</b> Use of default software/compute/network/dbParams profiles for database provisioning when software/compute/network/dbParams profiles are not passed

<br><b>Description :</b> This test case verifies that the database configured uses the default software/compute/network/dbParams profiles for configuration when software/compute/network/dbParams profiles are not present in the profiles section of "\ndb-operator\config\samples\ndb_v1alpha1_database.yaml"

<br><b>Pre-conditions :</b>
<ul>
<li>Pre-requisites are installed
<li>Docker Desktop Application is running
<li>Kubernetes cluster is up
<li>Nutanix Test Drive is active and the cluster id and other credentials are present inside "\ndb-operator\config\samples\ndb_v1alpha1_database.yaml" and "\ndb-operator\config\samples\secret.yaml"
<li>The software/compute/network/dbParams profiles are not available for input in a profiles section inside "\ndb-operator\config\samples\ndb_v1alpha1_database.yaml"
</ul>

<br><b>Test steps :</b>
<ul>
<li>Run command "make install run" in the root directory of the project
<li>Create secrets with command "kubectl apply -f .\config\samples\secret.yaml"
<li>Provision the database with command "kubectl apply -f .\config\samples\ndb_v1alpha1_database.yaml"
<li>Check if the appropriate database has been provisioned on the Nutanix test drive
<li>Verify that the compute/software/network/dbParams profiles of the database match the default profiles
</ul>

<br><b>Expected results :</b>
<ul>
<li>The system provisions the appropriate database based on the configurations specified in "\ndb-operator\config\samples\ndb_v1alpha1_database.yaml"
<li>The the compute/software/network/dbParams profiles match the default profile values
<li>The test case passes successfully
</ul>

<h2> Testing </h2>
Testcases were written in "\ndb-operator\test\ndb_api_helpers_test.go"
<br>Dummy Objects required for these testcases were created in "\ndb-operator\test\testutility.go"

<h3> Testcase to check Test Scenario 1 and Test Scenario 3 :</h3>

```
func TestEnrichAndGetProfilesWhenCustomProfilesMatch(t *testing.T) {

	//Set
	server := GetServerTestHelper(t)
	defer server.Close()
	ndbclient := ndbclient.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	dbTypes := []string{"postgres", "mysql", "mongodb"}

	for _, dbType := range dbTypes {

		// get custom profile based upon the database type
		customProfile := GetCustomProfileForDBType(dbType)

		profileMap, _ := v1alpha1.EnrichAndGetProfiles(context.Background(), ndbclient, dbType, customProfile)

		//Assert
		profileTypes := []string{
			v1alpha1.PROFILE_TYPE_COMPUTE,
			v1alpha1.PROFILE_TYPE_STORAGE,
			v1alpha1.PROFILE_TYPE_SOFTWARE,
			v1alpha1.PROFILE_TYPE_NETWORK,
			v1alpha1.PROFILE_TYPE_DATABASE_PARAMETER,
		}
		for _, profileType := range profileTypes {
			profile := profileMap[profileType]
			//Assert that no profileType is empty
			if profile == (v1alpha1.ProfileResponse{}) {
				t.Errorf("Empty profile type %s for dbType %s", profileType, dbType)
			}
			//Assert that profile EngineType matches the database engine or the generic type
			if profile.EngineType != v1alpha1.GetDatabaseEngineName(dbType) && profile.EngineType != v1alpha1.DATABASE_ENGINE_TYPE_GENERIC {
				t.Errorf("Profile engine type %s for dbType %s does not match", profile.EngineType, dbType)
			}
			obtainedProfile := v1alpha1.GetProfileForType(profileType, customProfile)
			// Ignoring Storage Profile Type as the Profile struct currently only supports compute, software, network and dbParam
			if profileType != v1alpha1.PROFILE_TYPE_STORAGE && profile.Id != obtainedProfile.Id && profile.LatestVersionId != obtainedProfile.VersionId {
				t.Errorf("Custom Profile Enrichment failed for profileType = %s and dbType = %s", profileType, dbType)
			}
		}
	}
}
```

<h4> Code for creating Dummy Objects required for this testcase :</h4>

```
func GetCustomProfileForDBType(dbType string) (profiles v1alpha1.Profiles) {
	switch dbType {
	case v1alpha1.DATABASE_TYPE_POSTGRES:
		profiles = v1alpha1.Profiles{
			// Custom Software Profile Name = "custom postgres software profile"
			Software: v1alpha1.Profile{
				Id:        "12",
				VersionId: "v-id-12",
			},
			// Custom ompute Name = "a"
			Compute: v1alpha1.Profile{
				Id:        "1",
				VersionId: "v-id-1",
			},
			Network: v1alpha1.Profile{
				Id:        "15",
				VersionId: "v-id-15",
			},
			DbParam: v1alpha1.Profile{
				Id:        "18",
				VersionId: "v-id-18",
			},
		}
		return profiles
```

<h3> Testcase to check Test Scenario 2 :</h3>

```
func TestEnrichAndGetProfilesWhenInvalidCustomProfilesProvided(t *testing.T) {

	//Set
	server := GetServerTestHelper(t)
	defer server.Close()
	ndbclient := ndbclient.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	dbTypes := []string{"postgres_invalid_profiles", "mysql_invalid_profiles", "mongodb_invalid_profiles"}

	for _, dbType := range dbTypes {

		// get custom profile based upon the database type
		customProfile := GetCustomProfileForDBType(dbType)

		profileMap, _ := v1alpha1.EnrichAndGetProfiles(context.Background(), ndbclient, dbType, customProfile)

		//Assert
		profileTypes := []string{
			v1alpha1.PROFILE_TYPE_COMPUTE,
			v1alpha1.PROFILE_TYPE_STORAGE,
			v1alpha1.PROFILE_TYPE_SOFTWARE,
			v1alpha1.PROFILE_TYPE_NETWORK,
			v1alpha1.PROFILE_TYPE_DATABASE_PARAMETER,
		}
		for _, profileType := range profileTypes {
			profile := profileMap[profileType]
			//Assert that profile EngineType matches the database engine or the generic type
			if profile.EngineType != v1alpha1.GetDatabaseEngineName(dbType) && profile.EngineType != v1alpha1.DATABASE_ENGINE_TYPE_GENERIC {
				t.Errorf("Profile engine type %s for dbType %s does not match", profile.EngineType, dbType)
			}
			/* since custom profile is passed it should not default to OOB, and err should be raised stating the custom profile passed does not exist,
			and thus database provisioning does not occur
			*/
			if profile != (v1alpha1.ProfileResponse{}) {
				t.Errorf("Incorrect Profile Match found for profile type = %s and dbType = %s", profileType, dbType)
			}
		}
	}
}
```


<h4> Code for creating Dummy Objects required for this testcase :</h4>

```
case v1alpha1.DATABASE_TYPE_MONGODB_INVALID_PROFILE, v1alpha1.DATABASE_TYPE_MYSQL_INVALID_PROFILE, v1alpha1.DATABASE_TYPE_POSTGRES_INVALID_PROFILE:
		// below custom profiles do not exist and will be used for the negative scenario
		profiles = v1alpha1.Profiles{
			Software: v1alpha1.Profile{
				Id:        "140",
				VersionId: "v-id-140",
			},
			Compute: v1alpha1.Profile{
				Id:        "100",
				VersionId: "v-id-100",
			},
			Network: v1alpha1.Profile{
				Id:        "170",
				VersionId: "v-id-170",
			},
			DbParam: v1alpha1.Profile{
				Id:        "200",
				VersionId: "v-id-200",
			},
		}
		return profiles
```

<h2>Github</h2>
<li> Repo: https://github.com/karan-47/ndb-operator/tree/feature/ntnx3


<h2>Mentors</h2>
<li> Prof. Edward F. Gehringer
<li> Krunal Jhaveri
<li> Manav Rajvanshi
<li> Krishna Saurabh Vankadaru
<li> Kartiki Bhandakkar

<h2>Contributors</h2>
<li> Karan Pradeep Gala (kgala2)
<li> Ashish Joshi (ajoshi24)
<li> Tilak Satra (trsatra)

<h2>References</h2>
[1] Nutanix. (n.d.). Nutanix Database Service. Retrieved from https://www.nutanix.com/products/database-service

[2] Kubernetes Operator Pattern https://kubernetes.io/docs/concepts/extend-kubernetes/operator

[3] NDB Operator Document - https://docs.google.com/document/d/1-VykKyIeky3n4JciIIrNgirk-Cn4pDT1behc9Yl8Nxk/

[4] Go Operator SDK - https://sdk.operatorframework.io/docs/buildingoperators/golang/tutorial/
