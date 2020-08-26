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
type GetCharactersCharacterIdContactsLabels200Ok struct {
	// label_id integer
	LabelId int64 `json:"label_id"`
	// label_name string
	LabelName string `json:"label_name"`
}
