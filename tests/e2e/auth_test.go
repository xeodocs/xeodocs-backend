package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestAuthFlows tests the complete authentication flow including user registration,
// password changes, and user cleanup. It validates that:
// - Admin users can register new non-admin users
// - Non-admin users can login with their credentials
// - Users can change their passwords
// - Users can login with updated passwords
// - Admin users can delete non-admin users
// - Deleted users cannot be accessed (404 response)
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

// TestUserCRUD tests the complete Create, Read, Update, Delete (CRUD) lifecycle
// of user resources through the admin interface. It validates that:
// - Admin users can register new users via /auth/register
// - Admin users can list all users
// - Admin users can retrieve individual users by ID
// - Admin users can update user details (username, role)
// - Admin users can delete users
// - Accessing deleted users returns 404 Not Found
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

// TestRoleCRUD tests the complete Create, Read, Update, Delete (CRUD) lifecycle
// of role resources through the admin interface. It validates that:
// - Admin users can create new roles with name and description
// - Admin users can list all roles (including default roles)
// - Admin users can retrieve individual roles by ID
// - Admin users can update role details (name, description)
// - Admin users can delete roles
// - Accessing deleted roles returns 404 Not Found
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
