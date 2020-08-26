/*
 * EVE Swagger Interface
 *
 * An OpenAPI for EVE Online
 *
 * API version: 1.7.2
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package esi

import (
	"time"
)

// 200 ok object
type GetCorporationsCorporationIdAlliancehistory200Ok struct {
	// alliance_id integer
	AllianceId int32 `json:"alliance_id,omitempty"`
	// True if the alliance has been closed
	IsDeleted bool `json:"is_deleted,omitempty"`
	// An incrementing ID that can be used to canonically establish order of records in cases where dates may be ambiguous
	RecordId int32 `json:"record_id"`
	// start_date string
	StartDate time.Time `json:"start_date"`
}
