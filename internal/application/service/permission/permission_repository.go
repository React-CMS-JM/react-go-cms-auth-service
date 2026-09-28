package permission

import (
	"context"

	"react-go-cms-auth-service/internal/domain/entity"
)

// Repository is the permission persistence port.
type Repository interface {
	// List returns every permission ordered by id.
	List(ctx context.Context) ([]entity.Permission, error)
}
