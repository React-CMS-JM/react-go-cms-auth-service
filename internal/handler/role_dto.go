package handler

import "react-go-cms-auth-service/internal/domain/entity"

// RoleResponse is the public role payload.
type RoleResponse struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	Permissions []string `json:"permissions"`
}

func toRoleResponse(role entity.Role) RoleResponse {
	return RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Permissions: emptyStrings(role.Permissions),
	}
}

func toRoleResponses(roles []entity.Role) []RoleResponse {
	if roles == nil {
		return []RoleResponse{}
	}
	var out []RoleResponse
	out = make([]RoleResponse, 0, len(roles))
	for _, role := range roles {
		out = append(out, toRoleResponse(role))
	}
	return out
}
