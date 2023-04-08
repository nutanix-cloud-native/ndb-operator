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

The codebase could be converted into an Object Oriented fashion with classes. Further, here are some of the design patterns we could use:

**Builder:** This pattern could be used to create the provisioning request for a database instance in a more modular and flexible way. Rather than creating the request directly in one function, a builder class could be used to set individual properties of the request. This would make the code more maintainable and extensible, and would allow for easier testing of different combinations of request properties.

**Factory:** Since there are different types of database instances that can be provisioned (e.g. MySQL, Postgres, etc.), a factory pattern could be used to create the appropriate request object based on the specified database type. This would help to decouple the creation of the request object from the calling code, and would make it easier to add support for new database types in the future.

**Dependency Injection:** To allow for better testability, dependency injection can be used to decouple the code from its dependencies. For example, in the provided code snippet, the NDBClient is being passed into the GenerateProvisioningRequest() function. However, if the NDBClient had additional dependencies or if it were difficult to create a testable version of the NDBClient, dependency injection could be used to allow for easier testing and swapping of dependencies.

**Code Refactoring:**
After reviewing the code base, it was discovered that the ndb_api_helpers.go file contains the code for provisioning the database. The main task of generating the request payload for provisioning the database is handled by the GenerateProvisioningRequest function. To retrieve all the profiles, this function utilizes the GetOOBProfiles function which returns a map of all the profiles. However, the current implementation of GenerateProvisioningRequest only retrieves the first element of the values within the map, which is assumed to be the default value. This means that the function doesn't verify if the user has provided a specific profile or not before assigning a default value.

To improve this behavior, we plan to iterate over all the profiles in the arrays that are inside the values of the profiles map. If the user has provided input for a specific profile, we will assign that input to the profiles variable. If not, we will use the first element of the array as the default value.

This change will allow us to properly check whether the user has provided input for a specific profile or not, and avoid the incorrect assumption that the first element of the values array is always the default value. By iterating over all the profiles, we can ensure that the correct profile is selected and assigned to the profiles variable. This will lead to more accurate and reliable database provisioning.

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
profilesMap, err := EnrichAndGetProfiles(ctx, ndbclient, dbSpec.Instance.Type, dbSpec.Instance.Profiles)
if err != nil {
    log.Error(err, "Error occurred while enriching and getting profiles", "database name", dbSpec.Instance.DatabaseInstanceName, "database type", dbSpec.Instance.Type)
    return
}
```

<br><b>Explanation for the change :</b>
changed the name of GetOOBProfiles to EnrichAndGetProfiles due to added functionality of overriding default profile values with custom profiles read from YAML file after performing applicability checks
</li>

<li><h4> EnrichAndGetProfiles </h4>

<b>Previous Working :</b> previously this function was named GetOOBProfiles. This function used to fetch all the profiles from NDB API and return ProfilesMap with default profiles for each of the compute, software, network and dbParams profiles.

<br><b>Enhanced Working :</b> now this function fetches all the profiles from NDB API and populates ProfilesMap with default profiles for each of the compute, software, network and dbParams profiles. Then it calls function EnrichProfilesMap function which will populate ProfilesMap with custom profiles if there are any in the YAML file.

<br><b>Previous Code :</b>
```
  profileMap[PROFILE_TYPE_COMPUTE] = computeProfiles[0]
	profileMap[PROFILE_TYPE_STORAGE] = storageProfiles[0]
	profileMap[PROFILE_TYPE_SOFTWARE] = softwareProfiles[0]
	profileMap[PROFILE_TYPE_NETWORK] = networkProfiles[0]
	profileMap[PROFILE_TYPE_DATABASE_PARAMETER] = dbParamProfiles[0]

	return
}
```

<br><b>New Code :</b>
```
  profileMap[PROFILE_TYPE_COMPUTE] = computeProfiles[0]
	profileMap[PROFILE_TYPE_STORAGE] = storageProfiles[0]
	profileMap[PROFILE_TYPE_SOFTWARE] = softwareProfiles[0]
	profileMap[PROFILE_TYPE_NETWORK] = networkProfiles[0]
	profileMap[PROFILE_TYPE_DATABASE_PARAMETER] = dbParamProfiles[0]

	// performs overriding of default OOB profiles based on the customProfiles obtained through YAML
	err = EnrichProfilesMap(ctx, customProfiles, genericProfiles, dbEngineSpecificProfiles, profileMap)

	return
}
```

<br><b>Explanation for the change :</b> since we only want to fall back to default profiles if there are no custom profiles mentioned in the YAML file, we are calling a new function EnrichProfilesMap which will populate ProfilesMap with the custom profiles.
</li>
<li><h4> EnrichProfilesMap </h4>

<b>Previous Working :</b> This function was not there previously.

<br><b>Enhanced Working :</b> This function checks if there are any custom profiles in the "\ndb-operator\config\samples\ndb_v1alpha1_database.yaml" file. If there any custom profiles, this function will call function PerformProfileMatchingAndEnrichProfiles to fetch them for each category (Compute, Software, Network, dbParams) and populate ProfilesMap with it.

<br><b>Previous Code :</b> N/A

<br><b>New Code :</b>
```
func EnrichProfilesMap(ctx context.Context, customProfiles Profiles, genericProfiles []ProfileResponse, dbEngineSpecificProfiles []ProfileResponse, profilesMap map[string]ProfileResponse) (err error) {
	log := ctrllog.FromContext(ctx)
	if customProfiles == (Profiles{}) {
		log.Info("Defaulting to using OOB Profiles as no custom profiles received. Returning from enrichingProfilesMap")
		return
	}
	log.Info("Received Custom Profiles => ", "Received Custom Profiles", customProfiles)
	customProfileOptions := [...]string{PROFILE_TYPE_COMPUTE, PROFILE_TYPE_SOFTWARE, PROFILE_TYPE_NETWORK, PROFILE_TYPE_DATABASE_PARAMETER}
	for _, profileValue := range customProfileOptions {
		err = PerformProfileMatchingAndEnrichProfiles(ctx, profileValue, customProfiles, genericProfiles, dbEngineSpecificProfiles, profilesMap)
		if err != nil {
			return
		}
	}
	return
}
```

<br><b>Explanation for the change :</b> Since we have added new section for custom profiles in the "\ndb-operator\config\samples\ndb_v1alpha1_database.yaml" file, we needed a function that will check if there is a section for custom profiles and delegate the task to fetch the custom profiles from the YAML file. This function fulfills that need.
</li>
<li><h4> PerformProfileMatchingAndEnrichProfiles </h4>

<b>Previous Working :</b> This function was not there previously.

<br><b>Enhanced Working :</b> Based on compute or (software, network & dbParam), generic or dbEngineSpecific profiles are used for matching the input customProfile. Furthermore, based on whether matched or not matched, delegation is performed to override the default profile values.

<br><b>Previous Code :</b> N/A

<br><b>New Code :</b>
```
func PerformProfileMatchingAndEnrichProfiles(ctx context.Context, profileType string, customProfiles Profiles, genericProfiles []ProfileResponse, dbEngineSpecificProfiles []ProfileResponse, profilesMap map[string]ProfileResponse) (err error) {
	log := ctrllog.FromContext(ctx)
	customProfile := GetProfileForType(profileType, customProfiles)
	if !isEmptyProfile(customProfile) {
		profileToUseForMatching := GetAppropriateProfileForType(profileType, genericProfiles, dbEngineSpecificProfiles)
		log.Info("Performing profile matching for profileType => ", "profileType", profileType)
		matchedProfile := util.Filter(profileToUseForMatching, func(p ProfileResponse) bool {
			return p.Type == profileType &&
				p.Id == customProfile.Id &&
				p.LatestVersionId == customProfile.VersionId &&
				p.Topology == GetTopologyForProfileType(profileType)
		})
		err = EnrichProfileMapForProfileType(ctx, profilesMap, profileType, matchedProfile)
	}
	return
}
```

<br><b>Explanation for the change :</b> We want to check if custom profiles mentioned in the YAML file are valid or not. If the profile type is compute, this function calls another function to validate the custom profile with generic profiles. If the profile type is network/software/dbParams, this function calls another function to validate the custom profile with dbEngineSpecific profiles.
</li>
<li><h4> GetAppropriateProfileForType </h4>

<b>Previous Working :</b> This function was not there previously.

<br><b>Enhanced Working :</b> This functions gives either generic or dbEngine specific profiles based upon the profile type to be filtered upon.

<br><b>Previous Code :</b> N/A

<br><b>New Code :</b>
```
func GetAppropriateProfileForType(profileType string, genericProfiles []ProfileResponse, dbEngineSpecificProfiles []ProfileResponse) (profiles []ProfileResponse) {
	if profileType == PROFILE_TYPE_COMPUTE {
		return genericProfiles
	} else {
		return dbEngineSpecificProfiles
	}
}
```

<br><b>Explanation for the change :</b> This function is used by PerformProfileMatchingAndEnrichProfiles function to make the decision of what kind of profiles are to be matched with what type of profiles.
</li>
<li><h4> EnrichProfileMapForProfileType </h4>

<b>Previous Working :</b> This function was not there previously.

<br><b>Enhanced Working :</b> This function checks the correctness of the profile (response) passed as the parameter and overrides the profilesMap for the custom profile type <br>specified if the custom profile provided passes the checks.

<b>Previous Code :</b> N/A

<br><b>New Code :</b>
```
func EnrichProfileMapForProfileType(ctx context.Context, profileMap map[string]ProfileResponse, profileType string, response []ProfileResponse) (err error) {
	log := ctrllog.FromContext(ctx)
	if len(response) == 0 {
		err = fmt.Errorf("No matching profile found for profileType = %s", profileType)
		log.Info("Error Occurred. No enrichment performed for profile => ", "profileType", profileType)
		return
	}
	log.Info("Profile Matching succeeded for profileType => ", "profileType", profileType)
	log.Info("Going to perform custom profile enrichment performed for => ", "profileType", profileType)
	profileMap[profileType] = response[0]
	return
}
```

<br><b>Explanation for the change :</b> The custom profile mentioned in the YAML file is only valid if it exists in the list of all profiles provided by the NDB API. This function performs the task to check if the given custom profile exists in the all profiles list.
</li>
<li><h4> GetTopologyForProfileType </h4>

<b>Previous Working :</b> This function was not there previously.

<br><b>Enhanced Working :</b> Providing the least costly topology based on each profile type.

<b>Previous Code :</b> N/A

<br><b>New Code :</b>
```
func GetTopologyForProfileType(profileType string) string {
	switch profileType {
	case PROFILE_TYPE_COMPUTE:
		return TOPOLOGY_ALL
	case PROFILE_TYPE_SOFTWARE:
		return TOPOLOGY_SINGLE
	case PROFILE_TYPE_NETWORK:
		return TOPOLOGY_ALL
	case PROFILE_TYPE_DATABASE_PARAMETER:
		return TOPOLOGY_INSTANCE
	default:
		return ""
	}
}
```

<br><b>Explanation for the change :</b> Costlier topologies need more space on NDB Test Drive which results in "NoHostResources" error. So this function chooses the least costly topology to avoid this error.
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
