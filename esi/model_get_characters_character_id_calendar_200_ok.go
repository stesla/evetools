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

// event
type GetCharactersCharacterIdCalendar200Ok struct {
	// event_date string
	EventDate time.Time `json:"event_date,omitempty"`
	// event_id integer
	EventId int32 `json:"event_id,omitempty"`
	// event_response string
	EventResponse string `json:"event_response,omitempty"`
	// importance integer
	Importance int32 `json:"importance,omitempty"`
	// title string
	Title string `json:"title,omitempty"`
}