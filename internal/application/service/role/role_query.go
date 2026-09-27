package role

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"react-go-cms-auth-service/internal/domain/apperror"
	"react-go-cms-auth-service/internal/domain/entity"
)

const msgRoleNotFound = "Role not found: "

// List returns every role.
func (s *Service) List(ctx context.Context) ([]entity.Role, error) {
	var roles []entity.Role
	var err error
	roles, err = s.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	if roles == nil {
		roles = []entity.Role{}
	}
	return roles, nil
}

// Get returns one role.
func (s *Service) Get(ctx context.Context, id int) (entity.Role, error) {
	var role entity.Role
	var err error
	role, err = s.repository.Get(ctx, id)
	if errorsIsNotFound(err) {
		return entity.Role{}, fmt.Errorf("get role: %w", apperror.NotFound(msgRoleNotFound+strconv.Itoa(id)))
	}
	if err != nil {
		return entity.Role{}, fmt.Errorf("get role: %w", err)
	}
	role.Permissions = emptyStrings(role.Permissions)
	return role, nil
}

func emptyStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func errorsIsNotFound(err error) bool {
	return errors.Is(err, apperror.ErrNotFound)
}
