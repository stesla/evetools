package sde

func Initialize(dir string) (err error) {
	if err = loadTypes(dir); err != nil {
		return
	}

	if err = loadGroups(dir, marketTypes); err != nil {
		return
	}

	if err = loadStations(dir); err != nil {
		return
	}

	if err = loadSolarSystems(dir); err != nil {
		return
	}

	if err = loadCorporations(dir); err != nil {
		return
	}

	return nil
}
