/*
 * EVE Swagger Interface
 *
 * An OpenAPI for EVE Online
 *
 * API version: 1.7.2
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package esi

// active_total object
type GetFwLeaderboardsCharactersActiveTotalActiveTotal struct {
	// Amount of kills
	Amount int32 `json:"amount,omitempty"`
	// character_id integer
	CharacterId int32 `json:"character_id,omitempty"`
}
