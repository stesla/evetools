/*
 * EVE Swagger Interface
 *
 * An OpenAPI for EVE Online
 *
 * API version: 1.7.2
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package esi

// 200 ok object
type GetIndustrySystems200Ok struct {
	// cost_indices array
	CostIndices []GetIndustrySystemsCostIndice `json:"cost_indices"`
	// solar_system_id integer
	SolarSystemId int32 `json:"solar_system_id"`
}