package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"react-go-cms-auth-service/internal/avatar"
	"react-go-cms-auth-service/internal/platform"
)

var emailRE = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

type Service struct {
	db       *sql.DB
	secret   string
	issuer   string
	keyID    string
	lifespan time.Duration
}

func New(db *sql.DB, cfg platform.Config) *Service {
	return &Service{
		db:       db,
		secret:   cfg.JWTSecret,
		issuer:   cfg.JWTIssuer,
		keyID:    cfg.JWTKeyID,
		lifespan: time.Duration(cfg.JWTLifespan) * time.Second,
	}
}

type User struct {
	ID          string               `json:"id"`
	Email       string               `json:"email"`
	FirstName   *string              `json:"firstName"`
	LastName    *string              `json:"lastName"`
	AvatarColor string               `json:"avatarColor"`
	IsBanned    bool                 `json:"isBanned"`
	BanReason   *string              `json:"banReason"`
	RoleIDs     []int                `json:"roleIds"`
	CreatedAt   platform.InstantJSON `json:"createdAt"`
	UpdatedAt   platform.InstantJSON `json:"updatedAt"`
}

type UserSummary struct {
	ID          string  `json:"id"`
	FirstName   *string `json:"firstName"`
	LastName    *string `json:"lastName"`
	AvatarColor string  `json:"avatarColor"`
}

type LoginResponse struct {
	Token       string   `json:"token"`
	TokenType   string   `json:"tokenType"`
	ExpiresIn   int64    `json:"expiresIn"`
	User        User     `json:"user"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

type MeResponse struct {
	User        User     `json:"user"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

type Role struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	Permissions []string `json:"permissions"`
}

type Permission struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type UserStats struct {
	Total  int64 `json:"total"`
	Banned int64 `json:"banned"`
}

func (s *Service) Login(ctx context.Context, email, password string) (LoginResponse, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return LoginResponse{}, platform.BadRequest("message", "email and password are required")
	}
	if !emailRE.MatchString(email) {
		return LoginResponse{}, platform.BadRequest("message", "must be a well-formed email address")
	}
	u, hash, err := s.loadByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LoginResponse{}, platform.Unauthorized("Invalid email or password")
		}
		return LoginResponse{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return LoginResponse{}, platform.Unauthorized("Invalid email or password")
	}
	if u.IsBanned {
		return LoginResponse{}, platform.Unauthorized("User is banned")
	}
	roles, perms, err := s.roleAndPermissionNames(ctx, u.ID)
	if err != nil {
		return LoginResponse{}, err
	}
	token, err := platform.SignHS256(s.secret, s.keyID, s.issuer, u.ID, u.Email, roles, perms, s.lifespan)
	if err != nil {
		return LoginResponse{}, err
	}
	return LoginResponse{
		Token:       token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.lifespan.Seconds()),
		User:        u,
		Roles:       roles,
		Permissions: perms,
	}, nil
}

func (s *Service) Me(ctx context.Context, userID string) (MeResponse, error) {
	u, err := s.loadByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return MeResponse{}, platform.NotFound("message", "User not found")
		}
		return MeResponse{}, err
	}
	roles, perms, err := s.roleAndPermissionNames(ctx, u.ID)
	if err != nil {
		return MeResponse{}, err
	}
	return MeResponse{User: u, Roles: roles, Permissions: perms}, nil
}

func (s *Service) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, email, first_name, last_name, is_banned, ban_reason, created_at, updated_at
		FROM users ORDER BY email`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := []User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := s.attachRoleIDs(ctx, users); err != nil {
		return nil, err
	}
	sort.SliceStable(users, func(i, j int) bool {
		return strings.ToLower(users[i].Email) < strings.ToLower(users[j].Email)
	})
	return users, nil
}

func (s *Service) Stats(ctx context.Context) (UserStats, error) {
	var st UserStats
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&st.Total); err != nil {
		return st, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE is_banned = 1`).Scan(&st.Banned); err != nil {
		return st, err
	}
	return st, nil
}

func (s *Service) UsersByIDs(ctx context.Context, ids []string) ([]UserSummary, error) {
	seen := map[string]struct{}{}
	ordered := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ordered = append(ordered, id)
	}
	out := []UserSummary{}
	if len(ordered) == 0 {
		return out, nil
	}
	for _, id := range ordered {
		var first, last sql.NullString
		var email string
		err := s.db.QueryRowContext(ctx, `SELECT email, first_name, last_name FROM users WHERE id = ?`, id).
			Scan(&email, &first, &last)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return nil, err
		}
		out = append(out, UserSummary{
			ID:          id,
			FirstName:   platform.StrPtr(first),
			LastName:    platform.StrPtr(last),
			AvatarColor: avatar.FromEmail(email),
		})
	}
	return out, nil
}

func (s *Service) GetUser(ctx context.Context, id string) (User, error) {
	u, err := s.loadByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, platform.NotFound("message", "User not found: "+id)
	}
	return u, err
}

type CreateUser struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	RoleIDs   []int  `json:"roleIds"`
}

func (s *Service) CreateUser(ctx context.Context, req CreateUser) (User, error) {
	if strings.TrimSpace(req.Email) == "" || !emailRE.MatchString(strings.TrimSpace(req.Email)) {
		return User{}, platform.BadRequest("message", "must be a well-formed email address")
	}
	if strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.FirstName) == "" || strings.TrimSpace(req.LastName) == "" {
		return User{}, platform.BadRequest("message", "password, firstName and lastName are required")
	}
	if req.RoleIDs == nil {
		return User{}, platform.BadRequest("message", "roleIds is required")
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	var exists int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM users WHERE LOWER(email) = ?`, email).Scan(&exists)
	if err == nil {
		return User{}, platform.BadRequest("message", "Email already in use")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return User{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	id := platform.NewUUID()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, is_banned, ban_reason)
		VALUES (?, ?, ?, ?, ?, 0, NULL)`,
		id, email, string(hash), strings.TrimSpace(req.FirstName), strings.TrimSpace(req.LastName)); err != nil {
		return User{}, err
	}
	if err := replaceRolesTx(ctx, tx, id, req.RoleIDs); err != nil {
		return User{}, err
	}
	if err := tx.Commit(); err != nil {
		return User{}, err
	}
	return s.GetUser(ctx, id)
}

type UpdateUser struct {
	Email     string  `json:"email"`
	Password  *string `json:"password"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	RoleIDs   []int   `json:"roleIds"`
}

func (s *Service) UpdateUser(ctx context.Context, id string, req UpdateUser) (User, error) {
	if _, err := s.GetUser(ctx, id); err != nil {
		return User{}, err
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || !emailRE.MatchString(email) {
		return User{}, platform.BadRequest("message", "must be a well-formed email address")
	}
	if strings.TrimSpace(req.FirstName) == "" || strings.TrimSpace(req.LastName) == "" {
		return User{}, platform.BadRequest("message", "firstName and lastName are required")
	}
	if req.RoleIDs == nil {
		return User{}, platform.BadRequest("message", "roleIds is required")
	}
	var other string
	err := s.db.QueryRowContext(ctx, `SELECT id FROM users WHERE LOWER(email) = ?`, email).Scan(&other)
	if err == nil && other != id {
		return User{}, platform.BadRequest("message", "Email already in use")
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return User{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()
	if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return User{}, err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE users SET email = ?, first_name = ?, last_name = ?, password_hash = ? WHERE id = ?`,
			email, strings.TrimSpace(req.FirstName), strings.TrimSpace(req.LastName), string(hash), id); err != nil {
			return User{}, err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `
			UPDATE users SET email = ?, first_name = ?, last_name = ? WHERE id = ?`,
			email, strings.TrimSpace(req.FirstName), strings.TrimSpace(req.LastName), id); err != nil {
			return User{}, err
		}
	}
	if err := replaceRolesTx(ctx, tx, id, req.RoleIDs); err != nil {
		return User{}, err
	}
	if err := tx.Commit(); err != nil {
		return User{}, err
	}
	return s.GetUser(ctx, id)
}

func (s *Service) Ban(ctx context.Context, id, reason string) (User, error) {
	if strings.TrimSpace(reason) == "" {
		return User{}, platform.BadRequest("message", "reason is required")
	}
	res, err := s.db.ExecContext(ctx, `UPDATE users SET is_banned = 1, ban_reason = ? WHERE id = ?`, reason, id)
	if err != nil {
		return User{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return User{}, platform.NotFound("message", "User not found: "+id)
	}
	return s.GetUser(ctx, id)
}

func (s *Service) Unban(ctx context.Context, id string) (User, error) {
	res, err := s.db.ExecContext(ctx, `UPDATE users SET is_banned = 0, ban_reason = NULL WHERE id = ?`, id)
	if err != nil {
		return User{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return User{}, platform.NotFound("message", "User not found: "+id)
	}
	return s.GetUser(ctx, id)
}

func (s *Service) ListRoles(ctx context.Context) ([]Role, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, description FROM roles ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	roles := []Role{}
	for rows.Next() {
		var r Role
		var desc sql.NullString
		if err := rows.Scan(&r.ID, &r.Name, &desc); err != nil {
			return nil, err
		}
		r.Description = platform.StrPtr(desc)
		r.Permissions = []string{}
		roles = append(roles, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	perms, err := s.permissionsByRole(ctx)
	if err != nil {
		return nil, err
	}
	for i := range roles {
		if p := perms[roles[i].ID]; p != nil {
			roles[i].Permissions = p
		}
	}
	return roles, nil
}

func (s *Service) GetRole(ctx context.Context, id int) (Role, error) {
	var r Role
	var desc sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT id, name, description FROM roles WHERE id = ?`, id).
		Scan(&r.ID, &r.Name, &desc)
	if errors.Is(err, sql.ErrNoRows) {
		return Role{}, platform.NotFound("message", "Role not found: "+itoa(id))
	}
	if err != nil {
		return Role{}, err
	}
	r.Description = platform.StrPtr(desc)
	r.Permissions = []string{}
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.name FROM permissions p
		JOIN roles_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = ? ORDER BY p.name`, id)
	if err != nil {
		return Role{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return Role{}, err
		}
		r.Permissions = append(r.Permissions, name)
	}
	return r, rows.Err()
}

func (s *Service) ListPermissions(ctx context.Context) ([]Permission, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, description FROM permissions ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Permission{}
	for rows.Next() {
		var p Permission
		var desc sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &desc); err != nil {
			return nil, err
		}
		p.Description = platform.StrPtr(desc)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Service) loadByEmail(ctx context.Context, email string) (User, string, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, first_name, last_name, is_banned, ban_reason, created_at, updated_at
		FROM users WHERE LOWER(email) = ?`, email)
	return scanUserHash(row)
}

func (s *Service) loadByID(ctx context.Context, id string) (User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, email, first_name, last_name, is_banned, ban_reason, created_at, updated_at
		FROM users WHERE id = ?`, id)
	u, err := scanUser(row)
	if err != nil {
		return User{}, err
	}
	users := []User{u}
	if err := s.attachRoleIDs(ctx, users); err != nil {
		return User{}, err
	}
	return users[0], nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanUser(row scannable) (User, error) {
	var u User
	var first, last, reason sql.NullString
	var created, updated sql.NullTime
	var banned sql.NullBool
	if err := row.Scan(&u.ID, &u.Email, &first, &last, &banned, &reason, &created, &updated); err != nil {
		return User{}, err
	}
	fillUser(&u, first, last, banned, reason, created, updated)
	return u, nil
}

func scanUserHash(row scannable) (User, string, error) {
	var u User
	var hash string
	var first, last, reason sql.NullString
	var created, updated sql.NullTime
	var banned sql.NullBool
	if err := row.Scan(&u.ID, &u.Email, &hash, &first, &last, &banned, &reason, &created, &updated); err != nil {
		return User{}, "", err
	}
	fillUser(&u, first, last, banned, reason, created, updated)
	return u, hash, nil
}

func fillUser(u *User, first, last sql.NullString, banned sql.NullBool, reason sql.NullString, created, updated sql.NullTime) {
	u.FirstName = platform.StrPtr(first)
	u.LastName = platform.StrPtr(last)
	u.IsBanned = banned.Valid && banned.Bool
	u.BanReason = platform.StrPtr(reason)
	u.CreatedAt = platform.InstantFrom(created)
	u.UpdatedAt = platform.InstantFrom(updated)
	u.AvatarColor = avatar.FromEmail(u.Email)
	u.RoleIDs = []int{}
}

func (s *Service) attachRoleIDs(ctx context.Context, users []User) error {
	if len(users) == 0 {
		return nil
	}
	idx := map[string]int{}
	ids := make([]any, len(users))
	for i, u := range users {
		idx[u.ID] = i
		ids[i] = u.ID
		users[i].RoleIDs = []int{}
	}
	q := `SELECT user_id, role_id FROM users_roles WHERE user_id IN (` + placeholders(len(ids)) + `) ORDER BY role_id`
	rows, err := s.db.QueryContext(ctx, q, ids...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var uid string
		var rid int
		if err := rows.Scan(&uid, &rid); err != nil {
			return err
		}
		if i, ok := idx[uid]; ok {
			users[i].RoleIDs = append(users[i].RoleIDs, rid)
		}
	}
	return rows.Err()
}

func (s *Service) roleAndPermissionNames(ctx context.Context, userID string) ([]string, []string, error) {
	roles := []string{}
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.name FROM roles r
		JOIN users_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = ? ORDER BY r.name`, userID)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return nil, nil, err
		}
		roles = append(roles, name)
	}
	rows.Close()
	perms := []string{}
	prows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT p.name FROM permissions p
		JOIN roles_permissions rp ON rp.permission_id = p.id
		JOIN users_roles ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = ? AND p.name IS NOT NULL
		ORDER BY p.name`, userID)
	if err != nil {
		return nil, nil, err
	}
	defer prows.Close()
	for prows.Next() {
		var name string
		if err := prows.Scan(&name); err != nil {
			return nil, nil, err
		}
		perms = append(perms, name)
	}
	return roles, perms, prows.Err()
}

func (s *Service) permissionsByRole(ctx context.Context) (map[int][]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT rp.role_id, p.name FROM roles_permissions rp
		JOIN permissions p ON p.id = rp.permission_id
		ORDER BY p.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int][]string{}
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		out[id] = append(out[id], name)
	}
	return out, rows.Err()
}

func replaceRolesTx(ctx context.Context, tx *sql.Tx, userID string, roleIDs []int) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM users_roles WHERE user_id = ?`, userID); err != nil {
		return err
	}
	seen := map[int]struct{}{}
	for _, id := range roleIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		var exists int
		err := tx.QueryRowContext(ctx, `SELECT 1 FROM roles WHERE id = ?`, id).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return platform.BadRequest("message", "Unknown role id: "+itoa(id))
		}
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO users_roles (user_id, role_id) VALUES (?, ?)`, userID, id); err != nil {
			return err
		}
	}
	return nil
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	b := strings.Repeat("?,", n)
	return b[:len(b)-1]
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [16]byte
	i := len(buf)
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

func StatusOf(err error) int {
	var api *platform.APIError
	if errors.As(err, &api) {
		return api.Status
	}
	return http.StatusInternalServerError
}
