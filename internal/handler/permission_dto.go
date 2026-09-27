package handler

import "react-go-cms-auth-service/internal/domain/entity"

// PermissionResponse is the public permission payload.
type PermissionResponse struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func toPermissionResponse(row entity.Permission) PermissionResponse {
	return PermissionResponse{ID: row.ID, Name: row.Name, Description: row.Description}
}

func toPermissionResponses(rows []entity.Permission) []PermissionResponse {
	if rows == nil {
		return []PermissionResponse{}
	}
	var out []PermissionResponse
	out = make([]PermissionResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toPermissionResponse(row))
	}
	return out
}
