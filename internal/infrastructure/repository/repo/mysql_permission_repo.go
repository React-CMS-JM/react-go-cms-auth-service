package repo

import (
	"context"
	"database/sql"

	"react-go-cms-auth-service/internal/application/service/permission"
	"react-go-cms-auth-service/internal/domain/entity"
	"react-go-cms-auth-service/internal/infrastructure/repository/model"
)

// PermissionRepository persists permissions in MySQL.
type PermissionRepository struct {
	db *sql.DB
}

// NewPermissionRepository builds a permission repository.
func NewPermissionRepository(db *sql.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

var _ permission.Repository = (*PermissionRepository)(nil)

// List returns every permission ordered by id.
func (r *PermissionRepository) List(ctx context.Context) ([]entity.Permission, error) {
	var rows *sql.Rows
	var err error
	rows, err = r.db.QueryContext(ctx, `SELECT id, name, description FROM permissions ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []entity.Permission
	out = []entity.Permission{}
	for rows.Next() {
		var row model.PermissionRow
		err = rows.Scan(&row.ID, &row.Name, &row.Description)
		if err != nil {
			return nil, err
		}
		out = append(out, entity.Permission{
			ID:          row.ID,
			Name:        row.Name,
			Description: StrPtr(row.Description),
		})
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return out, nil
}
