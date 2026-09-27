package handler

import (
	"errors"
	"net/http"

	"react-go-cms-auth-service/internal/application/service/permission"
	"react-go-cms-auth-service/internal/application/service/role"
	"react-go-cms-auth-service/internal/application/service/user"
	"react-go-cms-auth-service/internal/domain/apperror"
	"react-go-cms-auth-service/internal/infrastructure/httpx"
)

const (
	healthBody            = "auth-service-ok"
	errorField            = "message"
	invalidJSON           = "invalid JSON body"
	unauthorizedMessage   = "Unauthorized"
	internalServerMessage = "Internal server error"
)

// Handler exposes user, role, and permission HTTP endpoints.
type Handler struct {
	users       *user.Service
	roles       *role.Service
	permissions *permission.Service
	secret      string
	issuer      string
}

// New builds the HTTP handler.
func New(users *user.Service, roles *role.Service, permissions *permission.Service, secret, issuer string) *Handler {
	return &Handler{users: users, roles: roles, permissions: permissions, secret: secret, issuer: issuer}
}

// Register wires every auth service route.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/auth/health", h.health)
	mux.HandleFunc("POST /api/auth/login", h.login)
	mux.HandleFunc("POST /api/auth/logout", h.withAuth(h.logout))
	mux.HandleFunc("GET /api/auth/me", h.withAuth(h.me))

	mux.HandleFunc("GET /api/users/stats", h.withAuth(h.stats))
	mux.HandleFunc("GET /api/users/by-ids", h.withAuth(h.byIDs))
	mux.HandleFunc("GET /api/users/{id}", h.withAuth(h.getUser))
	mux.HandleFunc("PUT /api/users/{id}", h.withAuth(h.updateUser))
	mux.HandleFunc("POST /api/users/{id}/ban", h.withAuth(h.ban))
	mux.HandleFunc("POST /api/users/{id}/unban", h.withAuth(h.unban))
	mux.HandleFunc("GET /api/users", h.withAuth(h.listUsers))
	mux.HandleFunc("POST /api/users", h.withAuth(h.createUser))

	mux.HandleFunc("GET /api/roles/{id}", h.withAuth(h.getRole))
	mux.HandleFunc("GET /api/roles", h.withAuth(h.listRoles))
	mux.HandleFunc("GET /api/permissions", h.withAuth(h.listPermissions))
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte(healthBody))
}

func writeServiceError(w http.ResponseWriter, err error) {
	var domainErr *apperror.Error
	if errors.As(err, &domainErr) {
		var status int
		status = http.StatusBadRequest
		if errors.Is(err, apperror.ErrUnauthorized) {
			status = http.StatusUnauthorized
		} else if errors.Is(err, apperror.ErrNotFound) {
			status = http.StatusNotFound
		}
		httpx.WriteJSON(w, status, map[string]string{errorField: domainErr.Message})
		return
	}
	httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{errorField: internalServerMessage})
}
