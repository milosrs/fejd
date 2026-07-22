package authutil

import (
	"fejd-backend/auth"
	"fmt"
	"net/http"
)

func GetUserID(r *http.Request) (string, error) {
	userID := auth.GetUserIDFromRequest(r)
	if userID == "" {
		return "", fmt.Errorf("user not authenticated")
	}
	return userID, nil
}

func GetRoles(r *http.Request) []string {
	return auth.GetRolesFromRequest(r)
}

func HasRole(r *http.Request, role string) bool {
	for _, rl := range GetRoles(r) {
		if rl == role {
			return true
		}
	}
	return false
}
