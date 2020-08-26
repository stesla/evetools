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
type GetCorporationsCorporationIdShareholders200Ok struct {
	// share_count integer
	ShareCount int64 `json:"share_count"`
	// shareholder_id integer
	ShareholderId int32 `json:"shareholder_id"`
	// shareholder_type string
	ShareholderType string `json:"shareholder_type"`
}
