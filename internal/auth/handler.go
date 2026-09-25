package auth

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"react-go-cms-auth-service/internal/platform"
)

type Handler struct {
	svc    *Service
	secret string
	issuer string
}

func NewHandler(svc *Service, secret, issuer string) *Handler {
	return &Handler{svc: svc, secret: secret, issuer: issuer}
}

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
	_, _ = w.Write([]byte("auth-service-ok"))
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := platform.DecodeJSONLenient(r, &body); err != nil {
		platform.WriteError(w, platform.BadRequest("message", "invalid JSON body"), "message")
		return
	}
	res, err := h.svc.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		platform.WriteError(w, err, "message")
		return
	}
	platform.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	platform.WriteNoContent(w)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	res, err := h.svc.Me(r.Context(), subject(r))
	if err != nil {
		platform.WriteError(w, err, "message")
		return
	}
	platform.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.ListUsers(r.Context())
	if err != nil {
		platform.WriteError(w, err, "message")
		return
	}
	platform.WriteJSON(w, http.StatusOK, users)
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	st, err := h.svc.Stats(r.Context())
	if err != nil {
		platform.WriteError(w, err, "message")
		return
	}
	platform.WriteJSON(w, http.StatusOK, st)
}

func (h *Handler) byIDs(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Query().Get("ids")
	parts := []string{}
	if strings.TrimSpace(raw) != "" {
		for _, p := range strings.Split(raw, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				parts = append(parts, p)
			}
		}
	}
	rows, err := h.svc.UsersByIDs(r.Context(), parts)
	if err != nil {
		platform.WriteError(w, err, "message")
		return
	}
	platform.WriteJSON(w, http.StatusOK, rows)
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	u, err := h.svc.GetUser(r.Context(), r.PathValue("id"))
	if err != nil {
		platform.WriteError(w, err, "message")
		return
	}
	platform.WriteJSON(w, http.StatusOK, u)
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	var body CreateUser
	if err := platform.DecodeJSONLenient(r, &body); err != nil {
		platform.WriteError(w, platform.BadRequest("message", "invalid JSON body"), "message")
		return
	}
	u, err := h.svc.CreateUser(r.Context(), body)
	if err != nil {
		platform.WriteError(w, err, "message")
		return
	}
	platform.WriteJSON(w, http.StatusCreated, u)
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	var body UpdateUser
	if err := platform.DecodeJSONLenient(r, &body); err != nil {
		platform.WriteError(w, platform.BadRequest("message", "invalid JSON body"), "message")
		return
	}
	u, err := h.svc.UpdateUser(r.Context(), r.PathValue("id"), body)
	if err != nil {
		platform.WriteError(w, err, "message")
		return
	}
	platform.WriteJSON(w, http.StatusOK, u)
}

func (h *Handler) ban(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Reason string `json:"reason"`
	}
	if err := platform.DecodeJSONLenient(r, &body); err != nil {
		platform.WriteError(w, platform.BadRequest("message", "invalid JSON body"), "message")
		return
	}
	u, err := h.svc.Ban(r.Context(), r.PathValue("id"), body.Reason)
	if err != nil {
		platform.WriteError(w, err, "message")
		return
	}
	platform.WriteJSON(w, http.StatusOK, u)
}

func (h *Handler) unban(w http.ResponseWriter, r *http.Request) {
	u, err := h.svc.Unban(r.Context(), r.PathValue("id"))
	if err != nil {
		platform.WriteError(w, err, "message")
		return
	}
	platform.WriteJSON(w, http.StatusOK, u)
}

func (h *Handler) listRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.svc.ListRoles(r.Context())
	if err != nil {
		platform.WriteError(w, err, "message")
		return
	}
	platform.WriteJSON(w, http.StatusOK, roles)
}

func (h *Handler) getRole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		platform.WriteError(w, platform.NotFound("message", "Role not found: "+r.PathValue("id")), "message")
		return
	}
	role, err := h.svc.GetRole(r.Context(), id)
	if err != nil {
		platform.WriteError(w, err, "message")
		return
	}
	platform.WriteJSON(w, http.StatusOK, role)
}

func (h *Handler) listPermissions(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.ListPermissions(r.Context())
	if err != nil {
		platform.WriteError(w, err, "message")
		return
	}
	platform.WriteJSON(w, http.StatusOK, rows)
}

type ctxKey int

const subjectKey ctxKey = 1

func subject(r *http.Request) string {
	v, _ := r.Context().Value(subjectKey).(string)
	return v
}

func (h *Handler) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw := platform.BearerToken(r)
		if raw == "" {
			platform.WriteError(w, platform.Unauthorized("Unauthorized"), "message")
			return
		}
		claims, err := platform.ParseHS256(h.secret, h.issuer, raw)
		if err != nil || claims.Subject == "" {
			platform.WriteError(w, platform.Unauthorized("Unauthorized"), "message")
			return
		}
		ctx := context.WithValue(r.Context(), subjectKey, claims.Subject)
		next(w, r.WithContext(ctx))
	}
}
