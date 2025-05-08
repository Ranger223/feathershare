package sessionmiddleware

import (
	"context"
	"main/models"
	"net/http"
	"time"
)

type contextKey string

const ContextUserID = contextKey("userID")
const ContextIsAdmin = contextKey("isAdmin")

func SessionAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var userID int
		var isAdmin bool
		var expires time.Time
		row := models.DB.QueryRow(
			"SELECT u.id, u.isAdmin, s.expires_at FROM sessions s JOIN users u ON s.user_id = u.id WHERE s.id = ?", cookie.Value,
		)
		err = row.Scan(&userID, &isAdmin, &expires)
		if err != nil || expires.Before(time.Now()) {
			http.Error(w, "Session expired or invalid", http.StatusUnauthorized)
			return
		}

		// Add userID to context
		ctx := context.WithValue(r.Context(), ContextUserID, userID)
		ctx = context.WithValue(ctx, ContextIsAdmin, isAdmin)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
