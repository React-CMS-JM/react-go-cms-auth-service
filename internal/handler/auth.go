package handler

import (
	"context"
	"net/http"

	"react-go-cms-auth-service/internal/infrastructure/httpx"
	"react-go-cms-auth-service/internal/infrastructure/jwt"
)

type ctxKey int

const subjectKey ctxKey = 1

func subject(r *http.Request) string {
	var value string
	value, _ = r.Context().Value(subjectKey).(string)
	return value
}

func (h *Handler) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var raw string
		raw = httpx.BearerToken(r)
		if raw == "" {
			httpx.WriteError(w, httpx.Unauthorized(unauthorizedMessage), errorField)
			return
		}
		var claims *jwt.Claims
		var err error
		claims, err = jwt.ParseHS256(h.secret, h.issuer, raw)
		if err != nil || claims.Subject == "" {
			httpx.WriteError(w, httpx.Unauthorized(unauthorizedMessage), errorField)
			return
		}
		var ctx context.Context
		ctx = context.WithValue(r.Context(), subjectKey, claims.Subject)
		next(w, r.WithContext(ctx))
	}
}
