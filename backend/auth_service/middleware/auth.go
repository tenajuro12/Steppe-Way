package middleware

import (
	"authorization_service/utils/session"
	"context"
	"net/http"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, authenticated := utils.GetSessionUserID(r)
		if !authenticated {
			http.Error(w, "Unauthorized user", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
