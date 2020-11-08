package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/stesla/evetools/esd"
)

func usage() {
	program := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "USAGE: %s SDE.SQLITE3 > static.json", program)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	if err := esd.Initialize(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "error loading database file:", err)
		os.Exit(1)
	}

	groups, err := esd.GetMarketGroups()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error loading market groups:", err)
		os.Exit(1)
	}

	var roots = []int{}

	for _, g := range groups {
		if g.ParentID == nil {
			roots = append(roots, g.ID)
			continue
		}
		p := groups[*g.ParentID]
		p.Groups = append(p.Groups, g.ID)
	}

	types, err := esd.GetMarketTypes()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error loading market types:", err)
		os.Exit(1)
	}

	for _, t := range types {
		g := groups[t.GroupID]
		g.Types = append(g.Types, t.ID)
	}

	json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
		"root":   roots,
		"groups": groups,
		"types":  types,
	})
}
