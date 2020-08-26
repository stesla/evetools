/*
 * EVE Swagger Interface
 *
 * An OpenAPI for EVE Online
 *
 * API version: 1.7.2
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package esi

// participant object
type GetSovereigntyCampaignsParticipant struct {
	// alliance_id integer
	AllianceId int32 `json:"alliance_id"`
	// score number
	Score float32 `json:"score"`
}