package sde

var sdeDir string

func Initialize(dir string) (err error) {
	sdeDir = dir

	if err = loadTypes(dir); err != nil {
		return
	}

	return nil
}
