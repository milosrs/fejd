package keycloak

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ErrUserNotFound is returned by user lookups when no matching user exists.
var ErrUserNotFound = errors.New("keycloak user not found")

func (c *Client) SearchUsersByAttribute(ctx context.Context, attr, value string) ([]User, error) {
	params := url.Values{}
	params.Set("q", attr+":"+value)

	var users []User
	err := c.do(ctx, "GET", "/admin/realms/"+c.realm+"/users?"+params.Encode(), nil, &users)
	return users, err
}

// UpdateUserAttributes merges the given attributes into the user, preserving
// every other field of the user representation.
func (c *Client) UpdateUserAttributes(ctx context.Context, userID string, attrs map[string][]string) error {
	path := "/admin/realms/" + c.realm + "/users/" + userID

	var user map[string]any
	if err := c.do(ctx, "GET", path, nil, &user); err != nil {
		return err
	}

	attributes, _ := user["attributes"].(map[string]any)
	if attributes == nil {
		attributes = map[string]any{}
	}
	for k, v := range attrs {
		vals := make([]any, len(v))
		for i, s := range v {
			vals[i] = s
		}
		attributes[k] = vals
	}
	user["attributes"] = attributes

	return c.do(ctx, "PUT", path, user, nil)
}

// ListUsersByRequiredActionAndAge returns users carrying the required action
// whose creation timestamp is older than olderThan.
func (c *Client) ListUsersByRequiredActionAndAge(ctx context.Context, action string, olderThan time.Duration) ([]User, error) {
	var users []User
	if err := c.do(ctx, "GET", "/admin/realms/"+c.realm+"/users?max=1000", nil, &users); err != nil {
		return nil, err
	}
	return usersExpired(users, action, time.Now().UTC().Add(-olderThan)), nil
}

func (c *Client) DeleteUser(ctx context.Context, userID string) error {
	return c.do(ctx, "DELETE", "/admin/realms/"+c.realm+"/users/"+userID, nil, nil)
}

// GetUserByEmail returns the user whose email matches exactly, or
// ErrUserNotFound when there is no such user.
func (c *Client) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	params := url.Values{}
	params.Set("email", email)
	params.Set("exact", "true")

	var users []User
	if err := c.do(ctx, "GET", "/admin/realms/"+c.realm+"/users?"+params.Encode(), nil, &users); err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, ErrUserNotFound
	}
	return &users[0], nil
}

// AddRealmRole grants a realm role to a user by role name. It is idempotent:
// granting a role the user already holds is treated as success.
func (c *Client) AddRealmRole(ctx context.Context, userID, roleName string) error {
	path := "/admin/realms/" + c.realm + "/users/" + userID + "/role-mappings/realm"
	body := []map[string]string{{"name": roleName}}

	resp, err := c.request(ctx, "POST", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if roleMappingSuccess(resp.StatusCode) {
		return nil
	}
	respBody, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("keycloak admin api error: POST %s -> %d: %s", path, resp.StatusCode, string(respBody))
}

// ListRealmRoles returns the names of the realm roles mapped to a user.
func (c *Client) ListRealmRoles(ctx context.Context, userID string) ([]string, error) {
	var roles []struct {
		Name string `json:"name"`
	}
	if err := c.do(ctx, "GET", "/admin/realms/"+c.realm+"/users/"+userID+"/role-mappings/realm", nil, &roles); err != nil {
		return nil, err
	}

	names := make([]string, len(roles))
	for i, r := range roles {
		names[i] = r.Name
	}
	return names, nil
}

// roleMappingSuccess reports whether a role-mapping status code represents a
// successful grant (201 created, 204 no content, 409 already mapped).
func roleMappingSuccess(status int) bool {
	return status == http.StatusCreated || status == http.StatusNoContent || status == http.StatusConflict
}

// CreateUserInput describes a user to create via the admin API.
type CreateUserInput struct {
	Username        string
	Email           string
	FirstName       string
	EmailVerified   bool
	Enabled         bool
	RequiredActions []string
	Attributes      map[string][]string
}

// CreateUser creates a user and returns its ID (parsed from the Location
// header of the 201 response).
func (c *Client) CreateUser(ctx context.Context, in CreateUserInput) (string, error) {
	body := map[string]any{
		"username":        in.Username,
		"email":           in.Email,
		"emailVerified":   in.EmailVerified,
		"enabled":         in.Enabled,
		"requiredActions": in.RequiredActions,
		"attributes":      in.Attributes,
	}
	if in.FirstName != "" {
		body["firstName"] = in.FirstName
	}

	path := "/admin/realms/" + c.realm + "/users"
	resp, err := c.request(ctx, "POST", path, body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("keycloak admin api error: POST %s -> %d: %s", path, resp.StatusCode, string(respBody))
	}

	location := resp.Header.Get("Location")
	if i := strings.LastIndex(location, "/"); i >= 0 {
		return location[i+1:], nil
	}
	return "", fmt.Errorf("create user: no Location header in response")
}

// ExecuteActionsEmail sends an execute-actions email to the user for the given
// required actions, scoped to a client and redirect URI.
func (c *Client) ExecuteActionsEmail(ctx context.Context, userID string, actions []string, clientID, redirectURI string, lifespan int) error {
	params := url.Values{}
	if clientID != "" {
		params.Set("client_id", clientID)
	}
	if redirectURI != "" {
		params.Set("redirect_uri", redirectURI)
	}
	params.Set("lifespan", strconv.Itoa(lifespan))

	path := "/admin/realms/" + c.realm + "/users/" + userID + "/execute-actions-email?" + params.Encode()
	return c.do(ctx, "PUT", path, actions, nil)
}

// usersExpired filters users carrying the required action and created before
// cutoff. Pure so it can be unit-tested without a network round-trip.
func usersExpired(users []User, action string, cutoff time.Time) []User {
	var out []User
	for _, u := range users {
		if !containsString(u.RequiredActions, action) {
			continue
		}
		if u.CreatedTimestamp == 0 {
			continue
		}
		if time.UnixMilli(u.CreatedTimestamp).Before(cutoff) {
			out = append(out, u)
		}
	}
	return out
}

func containsString(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
