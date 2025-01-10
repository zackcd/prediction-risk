package core

import "fmt"

type ErrNotFound struct {
	Entity string
	ID     string
}

func (e *ErrNotFound) Error() string {
	if e.ID == "" {
		return fmt.Sprintf("%s not found", e.Entity)
	}
	return fmt.Sprintf("%s with ID %s not found", e.Entity, e.ID)
}

func NewErrNotFound(entity, id string) *ErrNotFound {
	return &ErrNotFound{Entity: entity, ID: id}
}
