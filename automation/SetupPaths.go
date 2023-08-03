package automation

type SetupPaths struct {
	dbSecretPath  string
	ndbSecretPath string
	dbPath        string
	appPodPath    string
	appSvcPath    string
}

/*
func (sp SetupPaths) createSecret(typ string) (err error) {
	if typ == "db" {
		createGeneric()
	} else if typ == "ndb" {

	} else {
		return errors.New("ERROR: Typ must be 'db' or 'ndb'!")
	}

}

func (sp SetupPaths) createDatabase() (database ndbv1alpha1.Database, err error) {
	generic, err := createGeneric(sp.dbPath, "database")
	if err != nil {
		return ndbv1alpha1.Database{}, err
	}

	val := reflect.ValueOf(generic)
	if val.Type() != reflect.TypeOf(ndbv1alpha1.Database{}) {
		return ndbv1alpha1.Database{}, err
	} else {
		database := val.Interface().(ndbv1alpha1.Database)
		return database, nil
	}
}

func (sp SetupPaths) createService() (secret v1.Service, err error) {
	generic, err := createGeneric(sp.appPodPath, "service")
	if err != nil {
		return v1.Service{}, err
	}

	val := reflect.ValueOf(generic)
	if val.Type() != reflect.TypeOf(v1.Service{}) {
		return v1.Service{}, err
	} else {
		service := val.Interface().(v1.Service)
		return service, nil
	}
}

func (sp SetupPaths) createPod() (secret v1.Pod, err error) {
	generic, err := createGeneric(sp.appPodPath, "pod")
	if err != nil {
		return v1.Pod{}, err
	}

	val := reflect.ValueOf(generic)
	if val.Type() != reflect.TypeOf(v1.Pod{}) {
		return v1.Pod{}, err
	} else {
		pod := val.Interface().(v1.Pod)
		return pod, nil
	}
}
*/
