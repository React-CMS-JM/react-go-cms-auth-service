package permission

import (
	"context"
	"fmt"

	"react-go-cms-auth-service/internal/domain/entity"
)

// List returns every permission.
func (s *Service) List(ctx context.Context) ([]entity.Permission, error) {
	var rows []entity.Permission
	var err error
	rows, err = s.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}
	if rows == nil {
		rows = []entity.Permission{}
	}
	return rows, nil
}
