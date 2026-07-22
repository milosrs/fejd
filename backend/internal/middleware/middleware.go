package middleware

import (
	"fejd-backend/internal/store"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func RequireBusinessAdmin(buStore *store.BusinessUserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			businessID, err := uuid.Parse(chi.URLParam(r, "businessID"))
			if err != nil {
				http.Error(w, `{"error":"invalid business ID"}`, http.StatusBadRequest)
				return
			}

			userID := getUserIDFromCtx(r)
			if userID == "" {
				http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
				return
			}

			isAdmin, err := buStore.IsAdmin(r.Context(), businessID, userID)
			if err != nil || !isAdmin {
				http.Error(w, `{"error":"forbidden: not admin of this business"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getUserIDFromCtx(r *http.Request) string {
	type ctxKey string
	if userID, ok := r.Context().Value(ctxKey("user_id")).(string); ok {
		return userID
	}
	return ""
}
