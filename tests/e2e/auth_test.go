package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthFlows(t *testing.T) {
	// 1. Login as default admin user
	adminClient := LoginAsAdmin(t)

	// 2. Change the password for the default admin user
	resp := adminClient.MakeRequest(t, "PUT", "/users/1", map[string]interface{}{
		"username": "admin",
		"password": "newpassword123",
	})
	adminClient.AssertSuccess(t, resp)
	var updatedUser map[string]interface{}
	adminClient.ParseJSON(t, resp, &updatedUser)
	require.Equal(t, "admin", updatedUser["username"])
	require.Equal(t, float64(1), updatedUser["id"]) // JSON numbers are float64

	// Verify login with new password
	newAdminClient := NewTestClient(gatewayURL)
	newAdminClient.Login(t, "admin", "newpassword123")

	// 3. Create a new non-admin user
	resp = newAdminClient.MakeRequest(t, "POST", "/auth/register", map[string]interface{}{
		"username": "testuser",
		"password": "testpass123",
		"role":     "viewer",
	})
	newAdminClient.AssertSuccess(t, resp)
	var registerRes map[string]interface{}
	newAdminClient.ParseJSON(t, resp, &registerRes)
	require.NotEmpty(t, registerRes["token"])

	// Get user id by listing users
	resp = newAdminClient.MakeRequest(t, "GET", "/users", nil)
	newAdminClient.AssertSuccess(t, resp)
	var users []map[string]interface{}
	newAdminClient.ParseJSON(t, resp, &users)
	var userID int
	for _, u := range users {
		if u["username"] == "testuser" {
			userID = int(u["id"].(float64))
			break
		}
	}
	require.NotZero(t, userID)

	// 4. Read the new non-admin user
	resp = newAdminClient.MakeRequest(t, "GET", fmt.Sprintf("/users/%d", userID), nil)
	newAdminClient.AssertSuccess(t, resp)
	var fetchedUser map[string]interface{}
	newAdminClient.ParseJSON(t, resp, &fetchedUser)
	require.Equal(t, "testuser", fetchedUser["username"])
	require.Equal(t, "viewer", fetchedUser["role"])

	// 5. Edit the new non-admin user
	resp = newAdminClient.MakeRequest(t, "PUT", fmt.Sprintf("/users/%d", userID), map[string]interface{}{
		"username": "updateduser",
		"role":     "editor",
	})
	newAdminClient.AssertSuccess(t, resp)
	var updatedUser2 map[string]interface{}
	newAdminClient.ParseJSON(t, resp, &updatedUser2)
	require.Equal(t, "updateduser", updatedUser2["username"])
	require.Equal(t, "editor", updatedUser2["role"])

	// 6. Delete the new non-admin user
	resp = newAdminClient.MakeRequest(t, "DELETE", fmt.Sprintf("/users/%d", userID), nil)
	newAdminClient.AssertNoContent(t, resp)

	// Verify deletion - should get 404
	resp = newAdminClient.MakeRequest(t, "GET", fmt.Sprintf("/users/%d", userID), nil)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestRoleCRUD(t *testing.T) {
	adminClient := LoginAsAdmin(t)

	// Create a new role
	resp := adminClient.MakeRequest(t, "POST", "/roles", map[string]interface{}{
		"name":        "moderator",
		"description": "Can moderate content",
	})
	adminClient.AssertSuccess(t, resp)
	var newRole map[string]interface{}
	adminClient.ParseJSON(t, resp, &newRole)
	roleID := int(newRole["id"].(float64))

	// List roles
	resp = adminClient.MakeRequest(t, "GET", "/roles", nil)
	adminClient.AssertSuccess(t, resp)
	var roles []map[string]interface{}
	adminClient.ParseJSON(t, resp, &roles)
	require.GreaterOrEqual(t, len(roles), 4) // default 3 + new one

	// Get specific role
	resp = adminClient.MakeRequest(t, "GET", fmt.Sprintf("/roles/%d", roleID), nil)
	adminClient.AssertSuccess(t, resp)
	var fetchedRole map[string]interface{}
	adminClient.ParseJSON(t, resp, &fetchedRole)
	require.Equal(t, "moderator", fetchedRole["name"])

	// Update role
	resp = adminClient.MakeRequest(t, "PUT", fmt.Sprintf("/roles/%d", roleID), map[string]interface{}{
		"name":        "supermod",
		"description": "Super moderator",
	})
	adminClient.AssertSuccess(t, resp)
	var updatedRole map[string]interface{}
	adminClient.ParseJSON(t, resp, &updatedRole)
	require.Equal(t, "supermod", updatedRole["name"])
	require.Equal(t, "Super moderator", updatedRole["description"])

	// Delete role
	resp = adminClient.MakeRequest(t, "DELETE", fmt.Sprintf("/roles/%d", roleID), nil)
	adminClient.AssertNoContent(t, resp)

	// Verify deletion
	resp = adminClient.MakeRequest(t, "GET", fmt.Sprintf("/roles/%d", roleID), nil)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
