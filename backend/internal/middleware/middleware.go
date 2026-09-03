package middleware

import (
	"fejd-backend/auth"
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
	return auth.GetUserIDFromRequest(r)
}

// MaxBodyBytes rejects requests whose body exceeds limit before any handler
// reads it. A declared Content-Length over the limit is rejected immediately;
// bodies without a known length (chunked) are wrapped with http.MaxBytesReader
// so reads are capped at the same limit.
func MaxBodyBytes(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > limit {
				http.Error(w, `{"error":"request body too large"}`, http.StatusRequestEntityTooLarge)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}
