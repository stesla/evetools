/*
 * EVE Swagger Interface
 *
 * An OpenAPI for EVE Online
 *
 * API version: 1.7.2
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package esi

// The aggressor corporation or alliance that declared this war, only contains either corporation_id or alliance_id
type GetWarsWarIdAggressor struct {
	// Alliance ID if and only if the aggressor is an alliance
	AllianceId int32 `json:"alliance_id,omitempty"`
	// Corporation ID if and only if the aggressor is a corporation
	CorporationId int32 `json:"corporation_id,omitempty"`
	// ISK value of ships the aggressor has destroyed
	IskDestroyed float32 `json:"isk_destroyed"`
	// The number of ships the aggressor has killed
	ShipsKilled int32 `json:"ships_killed"`
}
