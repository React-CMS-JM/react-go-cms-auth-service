package handler

import (
	"net/http"
	"strings"

	"react-go-cms-auth-service/internal/domain/apperror"
	"react-go-cms-auth-service/internal/domain/entity"
	"react-go-cms-auth-service/internal/infrastructure/httpx"
)

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var body LoginDTO
	var err error
	err = httpx.DecodeJSONLenient(r, &body)
	if err != nil {
		writeServiceError(w, apperror.Invalid(invalidJSON))
		return
	}
	var result entity.LoginResult
	result, err = h.users.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toLoginResponse(result))
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	httpx.WriteNoContent(w)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	var session entity.UserSession
	var err error
	session, err = h.users.Me(r.Context(), subject(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toMeResponse(session))
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	var users []entity.User
	var err error
	users, err = h.users.List(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toUserResponses(users))
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	var stats entity.UserStats
	var err error
	stats, err = h.users.Stats(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, UserStatsResponse{Total: stats.Total, Banned: stats.Banned})
}

func (h *Handler) byIDs(w http.ResponseWriter, r *http.Request) {
	var rows []entity.UserSummary
	var err error
	rows, err = h.users.ByIDs(r.Context(), queryIDs(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toUserSummaryResponses(rows))
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	var account entity.User
	var err error
	account, err = h.users.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toUserResponse(account))
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	var body CreateUserDTO
	var err error
	err = httpx.DecodeJSONLenient(r, &body)
	if err != nil {
		writeServiceError(w, apperror.Invalid(invalidJSON))
		return
	}
	var account entity.User
	account, err = h.users.Create(r.Context(), toUserCreate(body))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toUserResponse(account))
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	var body UpdateUserDTO
	var err error
	err = httpx.DecodeJSONLenient(r, &body)
	if err != nil {
		writeServiceError(w, apperror.Invalid(invalidJSON))
		return
	}
	var account entity.User
	account, err = h.users.Update(r.Context(), r.PathValue("id"), toUserUpdate(body))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toUserResponse(account))
}

func (h *Handler) ban(w http.ResponseWriter, r *http.Request) {
	var body BanDTO
	var err error
	err = httpx.DecodeJSONLenient(r, &body)
	if err != nil {
		writeServiceError(w, apperror.Invalid(invalidJSON))
		return
	}
	var account entity.User
	account, err = h.users.Ban(r.Context(), r.PathValue("id"), body.Reason)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toUserResponse(account))
}

func (h *Handler) unban(w http.ResponseWriter, r *http.Request) {
	var account entity.User
	var err error
	account, err = h.users.Unban(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toUserResponse(account))
}

func queryIDs(r *http.Request) []string {
	var raw string
	raw = r.URL.Query().Get("ids")
	var parts []string
	parts = []string{}
	if strings.TrimSpace(raw) == "" {
		return parts
	}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}
