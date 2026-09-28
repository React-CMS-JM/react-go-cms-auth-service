// Package user applies account use cases.
package user

import (
	"regexp"
	"time"
)

const (
	msgEmailAndPasswordRequired = "email and password are required"
	msgWellFormedEmail          = "must be a well-formed email address"
	msgInvalidCredentials       = "Invalid email or password"
	msgUserBanned               = "User is banned"
	msgUserNotFound             = "User not found"
	msgPasswordAndNameRequired  = "password, firstName and lastName are required"
	msgRoleIDsRequired          = "roleIds is required"
	msgEmailInUse               = "Email already in use"
	msgNameRequired             = "firstName and lastName are required"
	msgReasonRequired           = "reason is required"
	tokenTypeBearer             = "Bearer"
)

var emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// TokenIssuer signs a login access token.
type TokenIssuer interface {
	// Issue returns a signed token for the subject.
	Issue(subject, email string, roles, permissions []string, lifespan time.Duration) (string, error)
}

// Service coordinates user commands and queries.
type Service struct {
	repository Repository
	tokens     TokenIssuer
	lifespan   time.Duration
}

// New builds a user service. lifespan is the access-token lifetime.
func New(repository Repository, tokens TokenIssuer, lifespan time.Duration) *Service {
	return &Service{repository: repository, tokens: tokens, lifespan: lifespan}
}
