package user

import (
	"context"

	"react-go-cms-auth-service/internal/domain/entity"
)

// Repository is the user persistence port.
type Repository interface {
	// List returns every user ordered by the users query, without a case-insensitive sort.
	List(ctx context.Context) ([]entity.User, error)
	// Stats returns total and banned counts.
	Stats(ctx context.Context) (entity.UserStats, error)
	// FindSummary returns one short profile. The boolean is false when the id is missing.
	FindSummary(ctx context.Context, id string) (entity.UserSummary, bool, error)
	// Get returns one user with role ids. A missing id is ErrNotFound.
	Get(ctx context.Context, id string) (entity.User, error)
	// LoadByEmail returns the user and password hash. Role ids stay empty. A missing email is ErrNotFound.
	LoadByEmail(ctx context.Context, email string) (entity.User, string, error)
	// EmailTaken reports whether a lowercased email is already stored.
	EmailTaken(ctx context.Context, email string) (bool, error)
	// FindIDByEmail returns the user id for a lowercased email. The boolean is false when no row exists.
	FindIDByEmail(ctx context.Context, email string) (string, bool, error)
	// Insert stores a user and role links and returns the new id.
	Insert(ctx context.Context, email, passwordHash, firstName, lastName string, roleIDs []int) (string, error)
	// Update stores profile fields, an optional password hash, and replacement role links.
	Update(ctx context.Context, id, email, firstName, lastName string, passwordHash *string, roleIDs []int) error
	// Ban marks the user banned. A missing id is ErrNotFound.
	Ban(ctx context.Context, id, reason string) error
	// Unban clears the ban. A missing id is ErrNotFound.
	Unban(ctx context.Context, id string) error
	// RoleAndPermissionNames returns role names and distinct permission names for login and the caller profile.
	RoleAndPermissionNames(ctx context.Context, userID string) ([]string, []string, error)
}
