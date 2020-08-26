/*
 * EVE Swagger Interface
 *
 * An OpenAPI for EVE Online
 *
 * API version: 1.7.2
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package esi

// Optional object that is returned if a bookmark was made on a planet or a random location in space.
type GetCharactersCharacterIdBookmarksCoordinates struct {
	// x number
	X float64 `json:"x"`
	// y number
	Y float64 `json:"y"`
	// z number
	Z float64 `json:"z"`
}
