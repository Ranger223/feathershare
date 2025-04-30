package sessionmiddleware

import (
	"context"
	"main/models"
	"net/http"
	"time"
)

type contextKey string

const ContextUserID = contextKey("userID")

func SessionAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var userID int
		var expires time.Time
		row := models.DB.QueryRow(
			"SELECT user_id, expires_at From sessions WHERE id = ?", cookie.Value,
		)
		err = row.Scan(&userID, &expires)
		if err != nil || expires.Before(time.Now()) {
			http.Error(w, "Session expired or invalid", http.StatusUnauthorized)
			return
		}

		// Add userID to context
		ctx := context.WithValue(r.Context(), ContextUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
