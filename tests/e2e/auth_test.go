package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthFlows(t *testing.T) {
	// 1. Login as the default admin user
	adminClient := LoginAsAdmin(t)

	// 2. Create (register) a new non-admin user
	resp := adminClient.MakeRequest(t, "POST", "/auth/register", map[string]interface{}{
		"username": "testuser",
		"password": "oldpassword",
		"role":     "viewer",
	})
	adminClient.AssertSuccess(t, resp)
	var registerRes map[string]interface{}
	adminClient.ParseJSON(t, resp, &registerRes)
	require.NotEmpty(t, registerRes["token"])

	// Get user id by listing users
	resp = adminClient.MakeRequest(t, "GET", "/users", nil)
	adminClient.AssertSuccess(t, resp)
	var users []map[string]interface{}
	adminClient.ParseJSON(t, resp, &users)
	var userID int
	for _, u := range users {
		if u["username"] == "testuser" {
			userID = int(u["id"].(float64))
			break
		}
	}
	require.NotZero(t, userID)

	// 3. Login as the new non-admin user
	newUserClient := NewTestClient(gatewayURL)
	newUserClient.Login(t, "testuser", "oldpassword")

	// 4. Change the new user password
	resp = newUserClient.MakeRequest(t, "PUT", "/auth/change-password", map[string]interface{}{
		"password": "newpassword",
	})
	newUserClient.AssertNoContent(t, resp)

	// 5. Login as the new non-admin user using the new password
	newUserClientWithNewPass := NewTestClient(gatewayURL)
	newUserClientWithNewPass.Login(t, "testuser", "newpassword")

	// 6. Remove the new non-admin user
	resp = adminClient.MakeRequest(t, "DELETE", fmt.Sprintf("/users/%d", userID), nil)
	adminClient.AssertNoContent(t, resp)

	// Verify deletion - should get 404
	resp = adminClient.MakeRequest(t, "GET", fmt.Sprintf("/users/%d", userID), nil)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestUserCRUD(t *testing.T) {
	adminClient := LoginAsAdmin(t)

	// Create a new user
	resp := adminClient.MakeRequest(t, "POST", "/auth/register", map[string]interface{}{
		"username": "cruduser",
		"password": "crudpass",
		"role":     "viewer",
	})
	adminClient.AssertSuccess(t, resp)
	var registerRes map[string]interface{}
	adminClient.ParseJSON(t, resp, &registerRes)
	require.NotEmpty(t, registerRes["token"])

	// Get user id by listing users
	resp = adminClient.MakeRequest(t, "GET", "/users", nil)
	adminClient.AssertSuccess(t, resp)
	var users []map[string]interface{}
	adminClient.ParseJSON(t, resp, &users)
	var userID int
	for _, u := range users {
		if u["username"] == "cruduser" {
			userID = int(u["id"].(float64))
			break
		}
	}
	require.NotZero(t, userID)

	// Read the user
	resp = adminClient.MakeRequest(t, "GET", fmt.Sprintf("/users/%d", userID), nil)
	adminClient.AssertSuccess(t, resp)
	var fetchedUser map[string]interface{}
	adminClient.ParseJSON(t, resp, &fetchedUser)
	require.Equal(t, "cruduser", fetchedUser["username"])
	require.Equal(t, "viewer", fetchedUser["role"])

	// Update the user
	resp = adminClient.MakeRequest(t, "PUT", fmt.Sprintf("/users/%d", userID), map[string]interface{}{
		"username": "updatedcruduser",
		"role":     "editor",
	})
	adminClient.AssertSuccess(t, resp)
	var updatedUser map[string]interface{}
	adminClient.ParseJSON(t, resp, &updatedUser)
	require.Equal(t, "updatedcruduser", updatedUser["username"])
	require.Equal(t, "editor", updatedUser["role"])

	// Delete the user
	resp = adminClient.MakeRequest(t, "DELETE", fmt.Sprintf("/users/%d", userID), nil)
	adminClient.AssertNoContent(t, resp)

	// Verify deletion
	resp = adminClient.MakeRequest(t, "GET", fmt.Sprintf("/users/%d", userID), nil)
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
