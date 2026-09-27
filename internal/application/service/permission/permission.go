// Package permission applies permission queries.
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

// Service coordinates permission queries.
type Service struct {
	repository Repository
}

// New builds a permission service.
func New(repository Repository) *Service {
	return &Service{repository: repository}
}
