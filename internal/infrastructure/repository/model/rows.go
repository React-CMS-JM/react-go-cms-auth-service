// Package model holds database row shapes for the auth service.
package model

import "database/sql"

// UserRow is one users table row.
type UserRow struct {
	ID           string         `db:"id"`
	Email        string         `db:"email"`
	PasswordHash string         `db:"password_hash"`
	FirstName    sql.NullString `db:"first_name"`
	LastName     sql.NullString `db:"last_name"`
	IsBanned     sql.NullBool   `db:"is_banned"`
	BanReason    sql.NullString `db:"ban_reason"`
	CreatedAt    sql.NullTime   `db:"created_at"`
	UpdatedAt    sql.NullTime   `db:"updated_at"`
}

// UserSummaryRow is the short profile selected by id.
type UserSummaryRow struct {
	Email     string         `db:"email"`
	FirstName sql.NullString `db:"first_name"`
	LastName  sql.NullString `db:"last_name"`
}

// RoleRow is one roles table row.
type RoleRow struct {
	ID          int            `db:"id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
}

// PermissionRow is one permissions table row.
type PermissionRow struct {
	ID          int            `db:"id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
}
