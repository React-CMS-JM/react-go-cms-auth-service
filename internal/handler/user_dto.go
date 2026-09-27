package handler

import "react-go-cms-auth-service/internal/domain/entity"

// UserResponse is the public user payload.
type UserResponse struct {
	ID          string      `json:"id"`
	Email       string      `json:"email"`
	FirstName   *string     `json:"firstName"`
	LastName    *string     `json:"lastName"`
	AvatarColor string      `json:"avatarColor"`
	IsBanned    bool        `json:"isBanned"`
	BanReason   *string     `json:"banReason"`
	RoleIDs     []int       `json:"roleIds"`
	CreatedAt   InstantJSON `json:"createdAt"`
	UpdatedAt   InstantJSON `json:"updatedAt"`
}

// UserSummaryResponse is the short user payload.
type UserSummaryResponse struct {
	ID          string  `json:"id"`
	FirstName   *string `json:"firstName"`
	LastName    *string `json:"lastName"`
	AvatarColor string  `json:"avatarColor"`
}

// UserStatsResponse is the stats payload.
type UserStatsResponse struct {
	Total  int64 `json:"total"`
	Banned int64 `json:"banned"`
}

// LoginResponse is the login payload.
type LoginResponse struct {
	Token       string       `json:"token"`
	TokenType   string       `json:"tokenType"`
	ExpiresIn   int64        `json:"expiresIn"`
	User        UserResponse `json:"user"`
	Roles       []string     `json:"roles"`
	Permissions []string     `json:"permissions"`
}

// MeResponse is the caller payload.
type MeResponse struct {
	User        UserResponse `json:"user"`
	Roles       []string     `json:"roles"`
	Permissions []string     `json:"permissions"`
}

// LoginDTO is the login body.
type LoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreateUserDTO is the create-user body.
type CreateUserDTO struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	RoleIDs   []int  `json:"roleIds"`
}

// UpdateUserDTO is the update-user body.
type UpdateUserDTO struct {
	Email     string  `json:"email"`
	Password  *string `json:"password"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	RoleIDs   []int   `json:"roleIds"`
}

// BanDTO is the ban body.
type BanDTO struct {
	Reason string `json:"reason"`
}

func toUserResponse(account entity.User) UserResponse {
	return UserResponse{
		ID:          account.ID,
		Email:       account.Email,
		FirstName:   account.FirstName,
		LastName:    account.LastName,
		AvatarColor: account.AvatarColor,
		IsBanned:    account.IsBanned,
		BanReason:   account.BanReason,
		RoleIDs:     emptyInts(account.RoleIDs),
		CreatedAt:   toInstant(account.CreatedAt),
		UpdatedAt:   toInstant(account.UpdatedAt),
	}
}

func toUserResponses(users []entity.User) []UserResponse {
	if users == nil {
		return []UserResponse{}
	}
	var out []UserResponse
	out = make([]UserResponse, 0, len(users))
	for _, account := range users {
		out = append(out, toUserResponse(account))
	}
	return out
}

func toUserSummaryResponse(summary entity.UserSummary) UserSummaryResponse {
	return UserSummaryResponse{
		ID:          summary.ID,
		FirstName:   summary.FirstName,
		LastName:    summary.LastName,
		AvatarColor: summary.AvatarColor,
	}
}

func toUserSummaryResponses(rows []entity.UserSummary) []UserSummaryResponse {
	if rows == nil {
		return []UserSummaryResponse{}
	}
	var out []UserSummaryResponse
	out = make([]UserSummaryResponse, 0, len(rows))
	for _, summary := range rows {
		out = append(out, toUserSummaryResponse(summary))
	}
	return out
}

func toLoginResponse(result entity.LoginResult) LoginResponse {
	return LoginResponse{
		Token:       result.Token,
		TokenType:   result.TokenType,
		ExpiresIn:   result.ExpiresIn,
		User:        toUserResponse(result.User),
		Roles:       emptyStrings(result.Roles),
		Permissions: emptyStrings(result.Permissions),
	}
}

func toMeResponse(session entity.UserSession) MeResponse {
	return MeResponse{
		User:        toUserResponse(session.User),
		Roles:       emptyStrings(session.Roles),
		Permissions: emptyStrings(session.Permissions),
	}
}

func toUserCreate(body CreateUserDTO) entity.UserCreate {
	return entity.UserCreate{
		Email:     body.Email,
		Password:  body.Password,
		FirstName: body.FirstName,
		LastName:  body.LastName,
		RoleIDs:   body.RoleIDs,
	}
}

func toUserUpdate(body UpdateUserDTO) entity.UserUpdate {
	return entity.UserUpdate{
		Email:     body.Email,
		Password:  body.Password,
		FirstName: body.FirstName,
		LastName:  body.LastName,
		RoleIDs:   body.RoleIDs,
	}
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
