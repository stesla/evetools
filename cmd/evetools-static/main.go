package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	cfgFile          string
	corporationsFile string
	marketGroupsFile string
	marketTypesFile  string
	pkgName          string
	solarSystemsFile string
	stationsFile     string
)

func init() {
	pflag.StringVar(&cfgFile, "config", "", "config file, default: $HOME/.evetools.yaml")
	pflag.String("sde", "", "path to the SDE extract")
	viper.BindPFlag("sde.dir", pflag.Lookup("sde"))

	pflag.StringVar(&pkgName, "package", "sde", "package for generated files")
	pflag.StringVar(&corporationsFile, "corporations", "", "output corporations to this file")
	pflag.StringVar(&marketGroupsFile, "marketGroups", "", "output market groups to this file")
	pflag.StringVar(&marketTypesFile, "marketTypes", "", "output market types to this file")
	pflag.StringVar(&solarSystemsFile, "solarSystems", "", "output solar systems to this file")
	pflag.StringVar(&stationsFile, "stations", "", "output stations to this file")
}

func die(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}

func main() {
	pflag.Parse()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".evetools")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("$HOME")
	}
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("error loading config file: %s", err)
	}

	var sdeDir = viper.GetString("sde.dir")
	if sdeDir == "" {
		log.Fatalf("must set sde.dir")
	}

	if corporationsFile != "" {
		if err := buildCorporations(sdeDir, pkgName, corporationsFile); err != nil {
			die(err)
		}
	}

	if marketGroupsFile != "" {
		if err := buildMarketGroups(sdeDir, pkgName, marketGroupsFile); err != nil {
			die(err)
		}
	}

	if marketTypesFile != "" {
		if err := buildMarketTypes(sdeDir, pkgName, marketTypesFile); err != nil {
			die(err)
		}
	}

	if solarSystemsFile != "" {
		if err := buildSolarSystems(sdeDir, pkgName, solarSystemsFile); err != nil {
			die(err)
		}
	}

	if stationsFile != "" {
		if err := buildStations(sdeDir, pkgName, stationsFile); err != nil {
			die(err)
		}
	}
}
