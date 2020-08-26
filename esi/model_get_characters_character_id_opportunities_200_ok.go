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
type GetCharactersCharacterIdOpportunities200Ok struct {
	// completed_at string
	CompletedAt time.Time `json:"completed_at"`
	// task_id integer
	TaskId int32 `json:"task_id"`
}
