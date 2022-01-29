package types

import "github.com/google/uuid"

type Token struct {
	Common
	UserName string    `json:"username"`
	Token    uuid.UUID `json:"token"`
}
