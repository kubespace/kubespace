package types

import "github.com/google/uuid"

type UserSession struct {
	Common
	UserName  string    `json:"username"`
	SessionId uuid.UUID `json:"session_id"`
}
