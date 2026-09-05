package keycloak

import (
	"context"
	"net/url"
	"time"
)

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
