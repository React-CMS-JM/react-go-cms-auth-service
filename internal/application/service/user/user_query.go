package user

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"react-go-cms-auth-service/internal/domain/apperror"
	"react-go-cms-auth-service/internal/domain/avatar"
	"react-go-cms-auth-service/internal/domain/entity"
)

// Me returns the authenticated caller.
func (s *Service) Me(ctx context.Context, userID string) (entity.UserSession, error) {
	var account entity.User
	var err error
	account, err = s.repository.Get(ctx, userID)
	if errorsIsNotFound(err) {
		return entity.UserSession{}, fmt.Errorf("current user: %w", apperror.NotFound(msgUserNotFound))
	}
	if err != nil {
		return entity.UserSession{}, fmt.Errorf("current user: %w", err)
	}
	var roles []string
	var permissions []string
	roles, permissions, err = s.repository.RoleAndPermissionNames(ctx, account.ID)
	if err != nil {
		return entity.UserSession{}, fmt.Errorf("current user: %w", err)
	}
	return entity.UserSession{
		User:        applyAvatar(account),
		Roles:       emptyStrings(roles),
		Permissions: emptyStrings(permissions),
	}, nil
}

// List returns every user sorted by email, ignoring case.
func (s *Service) List(ctx context.Context) ([]entity.User, error) {
	var users []entity.User
	var err error
	users, err = s.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	if users == nil {
		users = []entity.User{}
	}
	sort.SliceStable(users, func(i, j int) bool {
		return strings.ToLower(users[i].Email) < strings.ToLower(users[j].Email)
	})
	for index := range users {
		users[index] = applyAvatar(users[index])
	}
	return users, nil
}

// Stats returns total and banned user counts.
func (s *Service) Stats(ctx context.Context) (entity.UserStats, error) {
	var stats entity.UserStats
	var err error
	stats, err = s.repository.Stats(ctx)
	if err != nil {
		return entity.UserStats{}, fmt.Errorf("user stats: %w", err)
	}
	return stats, nil
}

// ByIDs returns short profiles for the given ids, skipping blanks, duplicates, and missing rows.
func (s *Service) ByIDs(ctx context.Context, ids []string) ([]entity.UserSummary, error) {
	var seen map[string]struct{}
	seen = map[string]struct{}{}
	var ordered []string
	ordered = make([]string, 0, len(ids))
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
	var out []entity.UserSummary
	out = []entity.UserSummary{}
	if len(ordered) == 0 {
		return out, nil
	}
	for _, id := range ordered {
		var summary entity.UserSummary
		var found bool
		var err error
		summary, found, err = s.repository.FindSummary(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("users by ids: %w", err)
		}
		if !found {
			continue
		}
		summary.AvatarColor = avatar.FromEmail(summary.Email)
		out = append(out, summary)
	}
	return out, nil
}

// Get returns one user, including role ids.
func (s *Service) Get(ctx context.Context, id string) (entity.User, error) {
	var account entity.User
	var err error
	account, err = s.repository.Get(ctx, id)
	if errorsIsNotFound(err) {
		return entity.User{}, fmt.Errorf("get user: %w", apperror.NotFound(msgUserNotFound+": "+id))
	}
	if err != nil {
		return entity.User{}, fmt.Errorf("get user: %w", err)
	}
	return applyAvatar(account), nil
}

func applyAvatar(account entity.User) entity.User {
	account.AvatarColor = avatar.FromEmail(account.Email)
	account.RoleIDs = emptyInts(account.RoleIDs)
	return account
}

func emptyInts(values []int) []int {
	if values == nil {
		return []int{}
	}
	return values
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
