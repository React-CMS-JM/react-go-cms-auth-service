package repo

import (
	"context"
	"database/sql"
	"errors"

	"react-go-cms-auth-service/internal/application/service/role"
	"react-go-cms-auth-service/internal/domain/apperror"
	"react-go-cms-auth-service/internal/domain/entity"
	"react-go-cms-auth-service/internal/infrastructure/repository/model"
)

// RoleRepository persists roles in MySQL.
type RoleRepository struct {
	db *sql.DB
}

// NewRoleRepository builds a role repository.
func NewRoleRepository(db *sql.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

var _ role.Repository = (*RoleRepository)(nil)

// List returns every role with permission names.
func (r *RoleRepository) List(ctx context.Context) ([]entity.Role, error) {
	var rows *sql.Rows
	var err error
	rows, err = r.db.QueryContext(ctx, `SELECT id, name, description FROM roles ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []entity.Role
	roles = []entity.Role{}
	for rows.Next() {
		var row model.RoleRow
		err = rows.Scan(&row.ID, &row.Name, &row.Description)
		if err != nil {
			return nil, err
		}
		roles = append(roles, entity.Role{
			ID:          row.ID,
			Name:        row.Name,
			Description: StrPtr(row.Description),
			Permissions: []string{},
		})
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	var permissions map[int][]string
	permissions, err = r.permissionsByRole(ctx)
	if err != nil {
		return nil, err
	}
	for index := range roles {
		if names := permissions[roles[index].ID]; names != nil {
			roles[index].Permissions = names
		}
	}
	return roles, nil
}

// Get returns one role and its permission names.
func (r *RoleRepository) Get(ctx context.Context, id int) (entity.Role, error) {
	var row model.RoleRow
	var err error
	err = r.db.QueryRowContext(ctx, `SELECT id, name, description FROM roles WHERE id = ?`, id).
		Scan(&row.ID, &row.Name, &row.Description)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Role{}, apperror.ErrNotFound
	}
	if err != nil {
		return entity.Role{}, err
	}
	var role entity.Role
	role = entity.Role{
		ID:          row.ID,
		Name:        row.Name,
		Description: StrPtr(row.Description),
		Permissions: []string{},
	}
	var rows *sql.Rows
	rows, err = r.db.QueryContext(ctx, `
		SELECT p.name FROM permissions p
		JOIN roles_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = ? ORDER BY p.name`, id)
	if err != nil {
		return entity.Role{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return entity.Role{}, err
		}
		role.Permissions = append(role.Permissions, name)
	}
	err = rows.Err()
	if err != nil {
		return entity.Role{}, err
	}
	return role, nil
}

func (r *RoleRepository) permissionsByRole(ctx context.Context) (map[int][]string, error) {
	var rows *sql.Rows
	var err error
	rows, err = r.db.QueryContext(ctx, `
		SELECT rp.role_id, p.name FROM roles_permissions rp
		JOIN permissions p ON p.id = rp.permission_id
		ORDER BY p.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out map[int][]string
	out = map[int][]string{}
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, err
		}
		out[id] = append(out[id], name)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return out, nil
}
