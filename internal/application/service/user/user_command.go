package user

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"react-go-cms-auth-service/internal/domain/apperror"
	"react-go-cms-auth-service/internal/domain/avatar"
	"react-go-cms-auth-service/internal/domain/entity"
)

// Login checks the password and returns a signed token.
func (s *Service) Login(ctx context.Context, email, password string) (entity.LoginResult, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return entity.LoginResult{}, fmt.Errorf("login: %w", apperror.Invalid(msgEmailAndPasswordRequired))
	}
	if !emailPattern.MatchString(email) {
		return entity.LoginResult{}, fmt.Errorf("login: %w", apperror.Invalid(msgWellFormedEmail))
	}
	var account entity.User
	var passwordHash string
	var err error
	account, passwordHash, err = s.repository.LoadByEmail(ctx, email)
	if errorsIsNotFound(err) {
		return entity.LoginResult{}, fmt.Errorf("login: %w", apperror.Unauthorized(msgInvalidCredentials))
	}
	if err != nil {
		return entity.LoginResult{}, fmt.Errorf("login: %w", err)
	}
	var compareErr error
	compareErr = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if compareErr != nil {
		return entity.LoginResult{}, fmt.Errorf("login: %w", apperror.Unauthorized(msgInvalidCredentials))
	}
	if account.IsBanned {
		return entity.LoginResult{}, fmt.Errorf("login: %w", apperror.Unauthorized(msgUserBanned))
	}
	var roles []string
	var permissions []string
	roles, permissions, err = s.repository.RoleAndPermissionNames(ctx, account.ID)
	if err != nil {
		return entity.LoginResult{}, fmt.Errorf("login: %w", err)
	}
	var token string
	token, err = s.tokens.Issue(account.ID, account.Email, roles, permissions, s.lifespan)
	if err != nil {
		return entity.LoginResult{}, fmt.Errorf("login: %w", err)
	}
	account = applyAvatar(account)
	return entity.LoginResult{
		Token:       token,
		TokenType:   tokenTypeBearer,
		ExpiresIn:   int64(s.lifespan.Seconds()),
		User:        account,
		Roles:       emptyStrings(roles),
		Permissions: emptyStrings(permissions),
	}, nil
}

// Create stores a user and returns the saved account.
func (s *Service) Create(ctx context.Context, req entity.UserCreate) (entity.User, error) {
	var email string
	email = strings.TrimSpace(req.Email)
	if email == "" || !emailPattern.MatchString(email) {
		return entity.User{}, fmt.Errorf("create user: %w", apperror.Invalid(msgWellFormedEmail))
	}
	var password string
	password = strings.TrimSpace(req.Password)
	var firstName string
	firstName = strings.TrimSpace(req.FirstName)
	var lastName string
	lastName = strings.TrimSpace(req.LastName)
	if password == "" || firstName == "" || lastName == "" {
		return entity.User{}, fmt.Errorf("create user: %w", apperror.Invalid(msgPasswordAndNameRequired))
	}
	if req.RoleIDs == nil {
		return entity.User{}, fmt.Errorf("create user: %w", apperror.Invalid(msgRoleIDsRequired))
	}
	email = strings.ToLower(email)
	var taken bool
	var err error
	taken, err = s.repository.EmailTaken(ctx, email)
	if err != nil {
		return entity.User{}, fmt.Errorf("create user: %w", err)
	}
	if taken {
		return entity.User{}, fmt.Errorf("create user: %w", apperror.Invalid(msgEmailInUse))
	}
	var hash []byte
	hash, err = bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return entity.User{}, fmt.Errorf("create user: %w", err)
	}
	var id string
	id, err = s.repository.Insert(ctx, email, string(hash), firstName, lastName, req.RoleIDs)
	if err != nil {
		return entity.User{}, fmt.Errorf("create user: %w", err)
	}
	var created entity.User
	created, err = s.Get(ctx, id)
	if err != nil {
		return entity.User{}, err
	}
	created.AvatarColor = avatar.FromEmail(created.Email)
	return created, nil
}

// Update replaces profile fields and returns the saved account.
func (s *Service) Update(ctx context.Context, id string, req entity.UserUpdate) (entity.User, error) {
	var err error
	_, err = s.Get(ctx, id)
	if err != nil {
		return entity.User{}, err
	}
	var email string
	email = strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || !emailPattern.MatchString(email) {
		return entity.User{}, fmt.Errorf("update user: %w", apperror.Invalid(msgWellFormedEmail))
	}
	var firstName string
	firstName = strings.TrimSpace(req.FirstName)
	var lastName string
	lastName = strings.TrimSpace(req.LastName)
	if firstName == "" || lastName == "" {
		return entity.User{}, fmt.Errorf("update user: %w", apperror.Invalid(msgNameRequired))
	}
	if req.RoleIDs == nil {
		return entity.User{}, fmt.Errorf("update user: %w", apperror.Invalid(msgRoleIDsRequired))
	}
	var otherID string
	var found bool
	otherID, found, err = s.repository.FindIDByEmail(ctx, email)
	if err != nil {
		return entity.User{}, fmt.Errorf("update user: %w", err)
	}
	if found && otherID != id {
		return entity.User{}, fmt.Errorf("update user: %w", apperror.Invalid(msgEmailInUse))
	}
	var passwordHash *string
	if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
		var hash []byte
		hash, err = bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return entity.User{}, fmt.Errorf("update user: %w", err)
		}
		var encoded string
		encoded = string(hash)
		passwordHash = &encoded
	}
	err = s.repository.Update(ctx, id, email, firstName, lastName, passwordHash, req.RoleIDs)
	if err != nil {
		return entity.User{}, fmt.Errorf("update user: %w", err)
	}
	return s.Get(ctx, id)
}

// Ban stores a ban reason and returns the saved account.
func (s *Service) Ban(ctx context.Context, id, reason string) (entity.User, error) {
	if strings.TrimSpace(reason) == "" {
		return entity.User{}, fmt.Errorf("ban user: %w", apperror.Invalid(msgReasonRequired))
	}
	var err error
	err = s.repository.Ban(ctx, id, reason)
	if errorsIsNotFound(err) {
		return entity.User{}, fmt.Errorf("ban user: %w", apperror.NotFound(msgUserNotFound+": "+id))
	}
	if err != nil {
		return entity.User{}, fmt.Errorf("ban user: %w", err)
	}
	return s.Get(ctx, id)
}

// Unban clears a ban and returns the saved account.
func (s *Service) Unban(ctx context.Context, id string) (entity.User, error) {
	var err error
	err = s.repository.Unban(ctx, id)
	if errorsIsNotFound(err) {
		return entity.User{}, fmt.Errorf("unban user: %w", apperror.NotFound(msgUserNotFound+": "+id))
	}
	if err != nil {
		return entity.User{}, fmt.Errorf("unban user: %w", err)
	}
	return s.Get(ctx, id)
}
