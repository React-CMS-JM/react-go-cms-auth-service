package handler

import (
	"net/http"
	"strconv"

	"react-go-cms-auth-service/internal/domain/apperror"
	"react-go-cms-auth-service/internal/domain/entity"
	"react-go-cms-auth-service/internal/infrastructure/httpx"
)

const roleNotFoundPrefix = "Role not found: "

func (h *Handler) listRoles(w http.ResponseWriter, r *http.Request) {
	var roles []entity.Role
	var err error
	roles, err = h.roles.List(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toRoleResponses(roles))
}

func (h *Handler) getRole(w http.ResponseWriter, r *http.Request) {
	var rawID string
	rawID = r.PathValue("id")
	var id int
	var err error
	id, err = strconv.Atoi(rawID)
	if err != nil {
		writeServiceError(w, apperror.NotFound(roleNotFoundPrefix+rawID))
		return
	}
	var role entity.Role
	role, err = h.roles.Get(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toRoleResponse(role))
}
