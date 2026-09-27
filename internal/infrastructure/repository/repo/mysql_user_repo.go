package repo

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"react-go-cms-auth-service/internal/application/service/user"
	"react-go-cms-auth-service/internal/domain/apperror"
	"react-go-cms-auth-service/internal/domain/entity"
	"react-go-cms-auth-service/internal/infrastructure/identity"
	"react-go-cms-auth-service/internal/infrastructure/repository/model"
)

const msgUnknownRoleID = "Unknown role id: "

// UserRepository persists users in MySQL.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository builds a user repository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

var _ user.Repository = (*UserRepository)(nil)

type scannable interface {
	Scan(dest ...any) error
}

// List returns users in query order and attaches role ids.
func (r *UserRepository) List(ctx context.Context) ([]entity.User, error) {
	var rows *sql.Rows
	var err error
	rows, err = r.db.QueryContext(ctx, `
		SELECT id, email, first_name, last_name, is_banned, ban_reason, created_at, updated_at
		FROM users ORDER BY email`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []entity.User
	users = []entity.User{}
	for rows.Next() {
		var account entity.User
		account, err = scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, account)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	err = r.attachRoleIDs(ctx, users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// Stats returns total and banned counts.
func (r *UserRepository) Stats(ctx context.Context) (entity.UserStats, error) {
	var stats entity.UserStats
	var err error
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&stats.Total)
	if err != nil {
		return stats, err
	}
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE is_banned = 1`).Scan(&stats.Banned)
	if err != nil {
		return stats, err
	}
	return stats, nil
}

// FindSummary returns the short profile for one id.
func (r *UserRepository) FindSummary(ctx context.Context, id string) (entity.UserSummary, bool, error) {
	var row model.UserSummaryRow
	var err error
	err = r.db.QueryRowContext(ctx, `SELECT email, first_name, last_name FROM users WHERE id = ?`, id).
		Scan(&row.Email, &row.FirstName, &row.LastName)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.UserSummary{}, false, nil
	}
	if err != nil {
		return entity.UserSummary{}, false, err
	}
	return entity.UserSummary{
		ID:        id,
		Email:     row.Email,
		FirstName: StrPtr(row.FirstName),
		LastName:  StrPtr(row.LastName),
	}, true, nil
}

// Get returns one user with role ids.
func (r *UserRepository) Get(ctx context.Context, id string) (entity.User, error) {
	var row *sql.Row
	row = r.db.QueryRowContext(ctx, `
		SELECT id, email, first_name, last_name, is_banned, ban_reason, created_at, updated_at
		FROM users WHERE id = ?`, id)
	var account entity.User
	var err error
	account, err = scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.User{}, apperror.ErrNotFound
	}
	if err != nil {
		return entity.User{}, err
	}
	var users []entity.User
	users = []entity.User{account}
	err = r.attachRoleIDs(ctx, users)
	if err != nil {
		return entity.User{}, err
	}
	return users[0], nil
}

// LoadByEmail returns the user and password hash. Role ids stay empty.
func (r *UserRepository) LoadByEmail(ctx context.Context, email string) (entity.User, string, error) {
	var row *sql.Row
	row = r.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, first_name, last_name, is_banned, ban_reason, created_at, updated_at
		FROM users WHERE LOWER(email) = ?`, email)
	var account entity.User
	var passwordHash string
	var err error
	account, passwordHash, err = scanUserHash(row)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.User{}, "", apperror.ErrNotFound
	}
	if err != nil {
		return entity.User{}, "", err
	}
	return account, passwordHash, nil
}

// EmailTaken reports whether the lowercased email exists.
func (r *UserRepository) EmailTaken(ctx context.Context, email string) (bool, error) {
	var exists int
	var err error
	err = r.db.QueryRowContext(ctx, `SELECT 1 FROM users WHERE LOWER(email) = ?`, email).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// FindIDByEmail returns the id stored for a lowercased email.
func (r *UserRepository) FindIDByEmail(ctx context.Context, email string) (string, bool, error) {
	var id string
	var err error
	err = r.db.QueryRowContext(ctx, `SELECT id FROM users WHERE LOWER(email) = ?`, email).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return id, true, nil
}

// Insert stores a user and its roles.
func (r *UserRepository) Insert(ctx context.Context, email, passwordHash, firstName, lastName string, roleIDs []int) (string, error) {
	var id string
	id = identity.NewUUID()
	var tx *sql.Tx
	var err error
	tx, err = r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, is_banned, ban_reason)
		VALUES (?, ?, ?, ?, ?, 0, NULL)`,
		id, email, passwordHash, firstName, lastName)
	if err != nil {
		return "", err
	}
	err = replaceRoles(ctx, tx, id, roleIDs)
	if err != nil {
		return "", err
	}
	err = tx.Commit()
	if err != nil {
		return "", err
	}
	return id, nil
}

// Update stores profile fields and replacement roles.
func (r *UserRepository) Update(ctx context.Context, id, email, firstName, lastName string, passwordHash *string, roleIDs []int) error {
	var tx *sql.Tx
	var err error
	tx, err = r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if passwordHash != nil {
		_, err = tx.ExecContext(ctx, `
			UPDATE users SET email = ?, first_name = ?, last_name = ?, password_hash = ? WHERE id = ?`,
			email, firstName, lastName, *passwordHash, id)
		if err != nil {
			return err
		}
	} else {
		_, err = tx.ExecContext(ctx, `
			UPDATE users SET email = ?, first_name = ?, last_name = ? WHERE id = ?`,
			email, firstName, lastName, id)
		if err != nil {
			return err
		}
	}
	err = replaceRoles(ctx, tx, id, roleIDs)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// Ban sets the ban flag and reason.
func (r *UserRepository) Ban(ctx context.Context, id, reason string) error {
	var result sql.Result
	var err error
	result, err = r.db.ExecContext(ctx, `UPDATE users SET is_banned = 1, ban_reason = ? WHERE id = ?`, reason, id)
	if err != nil {
		return err
	}
	return missingWhenNone(result)
}

// Unban clears the ban flag and reason.
func (r *UserRepository) Unban(ctx context.Context, id string) error {
	var result sql.Result
	var err error
	result, err = r.db.ExecContext(ctx, `UPDATE users SET is_banned = 0, ban_reason = NULL WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return missingWhenNone(result)
}

// RoleAndPermissionNames returns role names and distinct permission names.
func (r *UserRepository) RoleAndPermissionNames(ctx context.Context, userID string) ([]string, []string, error) {
	var roles []string
	roles = []string{}
	var rows *sql.Rows
	var err error
	rows, err = r.db.QueryContext(ctx, `
		SELECT r.name FROM roles r
		JOIN users_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = ? ORDER BY r.name`, userID)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			rows.Close()
			return nil, nil, err
		}
		roles = append(roles, name)
	}
	rows.Close()
	var permissions []string
	permissions = []string{}
	var permissionRows *sql.Rows
	permissionRows, err = r.db.QueryContext(ctx, `
		SELECT DISTINCT p.name FROM permissions p
		JOIN roles_permissions rp ON rp.permission_id = p.id
		JOIN users_roles ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = ? AND p.name IS NOT NULL
		ORDER BY p.name`, userID)
	if err != nil {
		return nil, nil, err
	}
	defer permissionRows.Close()
	for permissionRows.Next() {
		var name string
		err = permissionRows.Scan(&name)
		if err != nil {
			return nil, nil, err
		}
		permissions = append(permissions, name)
	}
	err = permissionRows.Err()
	if err != nil {
		return nil, nil, err
	}
	return roles, permissions, nil
}

func (r *UserRepository) attachRoleIDs(ctx context.Context, users []entity.User) error {
	if len(users) == 0 {
		return nil
	}
	var indexByID map[string]int
	indexByID = map[string]int{}
	var ids []any
	ids = make([]any, len(users))
	for index, account := range users {
		indexByID[account.ID] = index
		ids[index] = account.ID
		users[index].RoleIDs = []int{}
	}
	var query string
	query = `SELECT user_id, role_id FROM users_roles WHERE user_id IN (` + placeholders(len(ids)) + `) ORDER BY role_id`
	var rows *sql.Rows
	var err error
	rows, err = r.db.QueryContext(ctx, query, ids...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var userID string
		var roleID int
		err = rows.Scan(&userID, &roleID)
		if err != nil {
			return err
		}
		if index, ok := indexByID[userID]; ok {
			users[index].RoleIDs = append(users[index].RoleIDs, roleID)
		}
	}
	return rows.Err()
}

func scanUser(row scannable) (entity.User, error) {
	var stored model.UserRow
	var err error
	err = row.Scan(&stored.ID, &stored.Email, &stored.FirstName, &stored.LastName, &stored.IsBanned, &stored.BanReason, &stored.CreatedAt, &stored.UpdatedAt)
	if err != nil {
		return entity.User{}, err
	}
	return userFromRow(stored), nil
}

func scanUserHash(row scannable) (entity.User, string, error) {
	var stored model.UserRow
	var err error
	err = row.Scan(&stored.ID, &stored.Email, &stored.PasswordHash, &stored.FirstName, &stored.LastName, &stored.IsBanned, &stored.BanReason, &stored.CreatedAt, &stored.UpdatedAt)
	if err != nil {
		return entity.User{}, "", err
	}
	return userFromRow(stored), stored.PasswordHash, nil
}

func userFromRow(row model.UserRow) entity.User {
	return entity.User{
		ID:        row.ID,
		Email:     row.Email,
		FirstName: StrPtr(row.FirstName),
		LastName:  StrPtr(row.LastName),
		IsBanned:  row.IsBanned.Valid && row.IsBanned.Bool,
		BanReason: StrPtr(row.BanReason),
		RoleIDs:   []int{},
		CreatedAt: InstantFrom(row.CreatedAt),
		UpdatedAt: InstantFrom(row.UpdatedAt),
	}
}

func replaceRoles(ctx context.Context, tx *sql.Tx, userID string, roleIDs []int) error {
	var err error
	_, err = tx.ExecContext(ctx, `DELETE FROM users_roles WHERE user_id = ?`, userID)
	if err != nil {
		return err
	}
	var seen map[int]struct{}
	seen = map[int]struct{}{}
	for _, id := range roleIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		var exists int
		err = tx.QueryRowContext(ctx, `SELECT 1 FROM roles WHERE id = ?`, id).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.Invalid(msgUnknownRoleID + itoa(id))
		}
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO users_roles (user_id, role_id) VALUES (?, ?)`, userID, id)
		if err != nil {
			return err
		}
	}
	return nil
}

func missingWhenNone(result sql.Result) error {
	var affected int64
	affected, _ = result.RowsAffected()
	if affected == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	var repeated string
	repeated = strings.Repeat("?,", n)
	return repeated[:len(repeated)-1]
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var neg bool
	neg = n < 0
	if neg {
		n = -n
	}
	var buf [16]byte
	var i int
	i = len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
