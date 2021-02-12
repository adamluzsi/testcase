package role

import (
	"context"
)

//----------------------------------------------------------------------------------------------------------------------
// package mydomain

type Entity struct {
	ID     string
	Field1 string
	Field2 int
	//...
}

type Consumer struct {
	Storage Storage
}

// role interface
type Storage interface {
	CreateEntity(ctx context.Context, ent *Entity) error
	FindEntityByID(ctx context.Context, ent *Entity, id string) (bool, error)
}

func (m Consumer) A(ctx context.Context, ent Entity) error {
	// validate entity

	if err := m.Storage.CreateEntity(ctx, &ent); err != nil {
		return err
	}

	// further steps here on entity

	return nil
}
