package handler

import (
	"net/http"

	"react-go-cms-auth-service/internal/domain/entity"
	"react-go-cms-auth-service/internal/infrastructure/httpx"
)

func (h *Handler) listPermissions(w http.ResponseWriter, r *http.Request) {
	var rows []entity.Permission
	var err error
	rows, err = h.permissions.List(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toPermissionResponses(rows))
}
