package sde

//go:generate evetools-static --config ../evetools.yaml --corporations corporations.go
type Corporation struct {
	ID        int    `json:"id"`
	Name      string `json:"string"`
	FactionID int    `json:"faction_id"`
}

//go:generate evetools-static --config ../evetools.yaml --marketTypes marketTypes.go
type MarketType struct {
	ID            int    `json:"id"`
	MarketGroupID int    `json:"market_group_id"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
}

//go:generate evetools-static --config ../evetools.yaml --marketGroups marketGroups.go
type MarketGroup struct {
	ID          int    `json:"id"`
	ParentID    int    `json:"parent_id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	Groups []int `json:"-"`
	Types  []int `json:"-"`
}

//go:generate evetools-static --config ../evetools.yaml --solarSystems solarSystems.go
type SolarSystem struct {
	ID       int    `json:"id" yaml:"solarSystemID"`
	Name     string `json:"name"`
	RegionID int    `json:"region_id" yaml:"regionID"`
}

//go:generate evetools-static --config ../evetools.yaml --stations stations.go
type Station struct {
	ID       int    `yaml:"stationID" json:"id"`
	Name     string `yaml:"stationName" json:"name"`
	CorpID   int    `yaml:"corporationID" json:"owner_id"`
	SystemID int    `yaml:"solarSystemID" json:"solar_system_id"`
	RegionID int    `yaml:"regionID" json:"region_id"`
}

var MarketGroupRoots []*MarketGroup

func init() {
	for id, g := range MarketGroups {
		if g.ParentID == 0 {
			MarketGroupRoots = append(MarketGroupRoots, g)
			continue
		}
		pg := MarketGroups[g.ParentID]
		pg.Groups = append(pg.Groups, id)
	}

	for id, t := range MarketTypes {
		if pg, found := MarketGroups[t.MarketGroupID]; found {
			pg.Types = append(pg.Types, id)
		}
	}
}

//go:generate gofmt -w .
