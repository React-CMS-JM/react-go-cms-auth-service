// Package entity holds auth-service domain objects without transport tags.
package entity

import "time"

// User is an account stored by the auth service.
type User struct {
	ID          string
	Email       string
	FirstName   *string
	LastName    *string
	AvatarColor string
	IsBanned    bool
	BanReason   *string
	RoleIDs     []int
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

// UserSummary is a short profile. Email chooses the avatar color and is omitted from the HTTP summary.
type UserSummary struct {
	ID          string
	Email       string
	FirstName   *string
	LastName    *string
	AvatarColor string
}

// UserStats is the admin user count.
type UserStats struct {
	Total  int64
	Banned int64
}

// UserCreate is the data required to insert a user.
type UserCreate struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	RoleIDs   []int
}

// UserUpdate replaces profile fields. A nil or blank Password leaves the hash unchanged.
type UserUpdate struct {
	Email     string
	Password  *string
	FirstName string
	LastName  string
	RoleIDs   []int
}

// LoginResult is a successful password login.
type LoginResult struct {
	Token       string
	TokenType   string
	ExpiresIn   int64
	User        User
	Roles       []string
	Permissions []string
}

// UserSession is the authenticated caller plus role and permission names.
type UserSession struct {
	User        User
	Roles       []string
	Permissions []string
}
