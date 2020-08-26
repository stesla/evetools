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
type GetCharactersCharacterIdOnlineOk struct {
	// Timestamp of the last login
	LastLogin time.Time `json:"last_login,omitempty"`
	// Timestamp of the last logout
	LastLogout time.Time `json:"last_logout,omitempty"`
	// Total number of times the character has logged in
	Logins int32 `json:"logins,omitempty"`
	// If the character is online
	Online bool `json:"online"`
}
