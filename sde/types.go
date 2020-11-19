package sde

type Corporation struct {
	ID        int    `json:"id"`
	Name      string `json:"string"`
	FactionID int    `json:"faction_id"`
}

type MarketGroup struct {
	ID          int    `json:"id"`
	ParentID    int    `json:"parent_id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`

	Groups []int `json:"groups"`
	Types  []int `json:"types"`
}

type MarketType struct {
	ID            int    `json:"id"`
	MarketGroupID int    `json:"market_group_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
}

type SolarSystem struct {
	ID       int    `json:"id" yaml:"solarSystemID"`
	Name     string `json:"name"`
	RegionID int    `json:"region_id" yaml:"regionID"`
}

type Station struct {
	ID       int    `yaml:"stationID" json:"id"`
	Name     string `yaml:"stationName" json:"name"`
	CorpID   int    `yaml:"corporationID" json:"corp_id"`
	RegionID int    `yaml:"regionID" json:"region_id"`
	SystemID int    `yaml:"solarSystemID" json:"system_id"`
}
