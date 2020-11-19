package sde

var sdeDir string

func Initialize(dir string) error {
	sdeDir = dir
	return nil
}
