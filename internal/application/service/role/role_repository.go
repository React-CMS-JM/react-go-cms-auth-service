package role

import (
	"context"

	"react-go-cms-auth-service/internal/domain/entity"
)

// Repository is the role persistence port.
type Repository interface {
	// List returns every role with permission names.
	List(ctx context.Context) ([]entity.Role, error)
	// Get returns one role. A missing id is ErrNotFound.
	Get(ctx context.Context, id int) (entity.Role, error)
}
