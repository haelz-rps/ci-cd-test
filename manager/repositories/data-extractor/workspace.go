package dataextractor

import "github.com/google/uuid"

type Workspace struct {
	ID   *uuid.UUID `json:"id"`
	Name string     `json:"name"`
}
