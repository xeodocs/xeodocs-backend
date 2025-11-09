package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// interfaceSliceToStringSlice converts a slice of interface{} to a slice of strings.
// This helper function is used in tests to compare JSON arrays returned from API
// responses with expected string arrays, since JSON unmarshaling produces interface{} slices.
func interfaceSliceToStringSlice(slice []interface{}) []string {
	result := make([]string, len(slice))
	for i, v := range slice {
		result[i] = v.(string)
	}
	return result
}

// TestProjectCRUD tests the complete Create, Read, Update, Delete (CRUD) lifecycle
// of project resources. It validates that:
// - Projects can be created with all required fields (name, repo_url, languages, build_commands)
// - Projects can be listed and the new project appears in the list
// - Individual projects can be retrieved by ID
// - Projects can be fully updated with new data (name, repo_url, languages, build_commands)
// - The updated_at timestamp changes after updates
// - Projects can be deleted
// - Accessing deleted projects returns 404 Not Found
func TestProjectCRUD(t *testing.T) {
	adminClient := LoginAsAdmin(t)

	// Create a new project
	projectData := map[string]interface{}{
		"name":           "Test Documentation Project",
		"doc_url":        "https://example.com/docs",
		"repo_url":       "https://github.com/example/docs",
		"languages":      []string{"en", "es", "fr"},
		"build_command":  "npm run build",
		"export_command": "npm run export",
		"preview_command": "npm run preview",
	}

	resp := adminClient.MakeRequest(t, "POST", "/projects", projectData)
	adminClient.AssertCreated(t, resp)
	var createdProject map[string]interface{}
	adminClient.ParseJSON(t, resp, &createdProject)
	require.NotEmpty(t, createdProject["id"])
	require.Equal(t, projectData["name"], createdProject["name"])
	require.Equal(t, projectData["doc_url"], createdProject["doc_url"])
	require.Equal(t, projectData["repo_url"], createdProject["repo_url"])
	require.Equal(t, projectData["languages"], interfaceSliceToStringSlice(createdProject["languages"].([]interface{})))
	require.Equal(t, projectData["build_command"], createdProject["build_command"])
	require.Equal(t, projectData["export_command"], createdProject["export_command"])
	require.Equal(t, projectData["preview_command"], createdProject["preview_command"])
	require.NotEmpty(t, createdProject["created_at"])
	require.NotEmpty(t, createdProject["updated_at"])

	projectID := int(createdProject["id"].(float64))

	// List projects - should include the newly created project
	resp = adminClient.MakeRequest(t, "GET", "/projects", nil)
	adminClient.AssertSuccess(t, resp)
	var projects []map[string]interface{}
	adminClient.ParseJSON(t, resp, &projects)
	require.GreaterOrEqual(t, len(projects), 1)

	// Find our project in the list
	var foundProject map[string]interface{}
	for _, p := range projects {
		if int(p["id"].(float64)) == projectID {
			foundProject = p
			break
		}
	}
	require.NotNil(t, foundProject)
	require.Equal(t, projectData["name"], foundProject["name"])

	// Get specific project by ID
	resp = adminClient.MakeRequest(t, "GET", fmt.Sprintf("/projects/%d", projectID), nil)
	adminClient.AssertSuccess(t, resp)
	var fetchedProject map[string]interface{}
	adminClient.ParseJSON(t, resp, &fetchedProject)
	require.Equal(t, projectID, int(fetchedProject["id"].(float64)))
	require.Equal(t, projectData["name"], fetchedProject["name"])
	require.Equal(t, projectData["doc_url"], fetchedProject["doc_url"])
	require.Equal(t, projectData["repo_url"], fetchedProject["repo_url"])
	require.Equal(t, projectData["languages"], interfaceSliceToStringSlice(fetchedProject["languages"].([]interface{})))
	require.Equal(t, projectData["build_command"], fetchedProject["build_command"])
	require.Equal(t, projectData["export_command"], fetchedProject["export_command"])
	require.Equal(t, projectData["preview_command"], fetchedProject["preview_command"])

	// Update the project
	updatedData := map[string]interface{}{
		"name":           "Updated Documentation Project",
		"doc_url":        "https://example.com/updated-docs",
		"repo_url":       "https://github.com/example/updated-docs",
		"languages":      []string{"en", "es", "fr", "de"},
		"build_command":  "yarn build",
		"export_command": "yarn export",
		"preview_command": "yarn preview",
	}

	resp = adminClient.MakeRequest(t, "PUT", fmt.Sprintf("/projects/%d", projectID), updatedData)
	adminClient.AssertSuccess(t, resp)
	var updatedProject map[string]interface{}
	adminClient.ParseJSON(t, resp, &updatedProject)
	require.Equal(t, projectID, int(updatedProject["id"].(float64)))
	require.Equal(t, updatedData["name"], updatedProject["name"])
	require.Equal(t, updatedData["doc_url"], updatedProject["doc_url"])
	require.Equal(t, updatedData["repo_url"], updatedProject["repo_url"])
	require.Equal(t, updatedData["languages"], interfaceSliceToStringSlice(updatedProject["languages"].([]interface{})))
	require.Equal(t, updatedData["build_command"], updatedProject["build_command"])
	require.Equal(t, updatedData["export_command"], updatedProject["export_command"])
	require.Equal(t, updatedData["preview_command"], updatedProject["preview_command"])
	require.NotEqual(t, createdProject["updated_at"], updatedProject["updated_at"])

	// Delete the project
	resp = adminClient.MakeRequest(t, "DELETE", fmt.Sprintf("/projects/%d", projectID), nil)
	adminClient.AssertNoContent(t, resp)

	// Verify deletion - should get 404
	resp = adminClient.MakeRequest(t, "GET", fmt.Sprintf("/projects/%d", projectID), nil)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestProjectSharedAccess tests that project resources are accessible to all authenticated users.
// It ensures that:
// - Any authenticated user can access projects created by others
// - Users can list all projects regardless of creator
// - Proper cleanup of test data (projects and users)
func TestProjectSharedAccess(t *testing.T) {
	adminClient := LoginAsAdmin(t)

	// Create a project as admin
	projectData := map[string]interface{}{
		"name":           "Shared Project",
		"doc_url":        "https://example.com/shared-docs",
		"repo_url":       "https://github.com/shared/docs",
		"languages":      []string{"en"},
		"build_command":  "npm install",
		"export_command": "npm run export",
		"preview_command": "npm run preview",
	}

	resp := adminClient.MakeRequest(t, "POST", "/projects", projectData)
	adminClient.AssertCreated(t, resp)
	var adminProject map[string]interface{}
	adminClient.ParseJSON(t, resp, &adminProject)
	adminProjectID := int(adminProject["id"].(float64))

	// Create a new user
	username := fmt.Sprintf("testuser%d", time.Now().Unix())
	resp = adminClient.MakeRequest(t, "POST", "/auth/register", map[string]interface{}{
		"username": username,
		"password": "testpass",
		"role":     "viewer",
	})
	adminClient.AssertSuccess(t, resp)

	// Login as the new user
	newUserClient := NewTestClient(gatewayURL)
	newUserClient.Login(t, username, "testpass")

	// User should be able to access admin's project
	resp = newUserClient.MakeRequest(t, "GET", fmt.Sprintf("/projects/%d", adminProjectID), nil)
	newUserClient.AssertSuccess(t, resp)
	var fetchedProject map[string]interface{}
	newUserClient.ParseJSON(t, resp, &fetchedProject)
	require.Equal(t, projectData["name"], fetchedProject["name"])

	// User should see the project in their list
	resp = newUserClient.MakeRequest(t, "GET", "/projects", nil)
	newUserClient.AssertSuccess(t, resp)
	var userProjects []map[string]interface{}
	newUserClient.ParseJSON(t, resp, &userProjects)

	// Should see admin's project
	projectFound := false
	for _, p := range userProjects {
		if int(p["id"].(float64)) == adminProjectID {
			projectFound = true
			break
		}
	}
	require.True(t, projectFound, "User should see shared project")

	// Create a project as the new user
	userProjectData := map[string]interface{}{
		"name":           "User Shared Project",
		"doc_url":        "https://example.com/user-docs",
		"repo_url":       "https://github.com/user/docs",
		"languages":      []string{"en", "fr"},
		"build_command":  "yarn install",
		"export_command": "yarn export",
		"preview_command": "yarn preview",
	}

	resp = newUserClient.MakeRequest(t, "POST", "/projects", userProjectData)
	newUserClient.AssertCreated(t, resp)
	var userProject map[string]interface{}
	newUserClient.ParseJSON(t, resp, &userProject)
	userProjectID := int(userProject["id"].(float64))

	// Admin should be able to access user's project
	resp = adminClient.MakeRequest(t, "GET", fmt.Sprintf("/projects/%d", userProjectID), nil)
	adminClient.AssertSuccess(t, resp)
	var fetchedUserProject map[string]interface{}
	adminClient.ParseJSON(t, resp, &fetchedUserProject)
	require.Equal(t, userProjectData["name"], fetchedUserProject["name"])

	// Clean up - delete the projects
	resp = adminClient.MakeRequest(t, "DELETE", fmt.Sprintf("/projects/%d", adminProjectID), nil)
	adminClient.AssertNoContent(t, resp)

	resp = newUserClient.MakeRequest(t, "DELETE", fmt.Sprintf("/projects/%d", userProjectID), nil)
	newUserClient.AssertNoContent(t, resp)

	// Clean up - delete the test user
	// Get user id by listing users
	resp = adminClient.MakeRequest(t, "GET", "/users", nil)
	adminClient.AssertSuccess(t, resp)
	var users []map[string]interface{}
	adminClient.ParseJSON(t, resp, &users)
	var userID int
	for _, u := range users {
		if u["username"] == username {
			userID = int(u["id"].(float64))
			break
		}
	}
	require.NotZero(t, userID)

	resp = adminClient.MakeRequest(t, "DELETE", fmt.Sprintf("/users/%d", userID), nil)
	adminClient.AssertNoContent(t, resp)
}

// TestProjectPartialUpdate tests the ability to update only specific fields of a project
// without affecting other fields. It validates that:
// - Only the provided fields are updated (partial updates)
// - Unspecified fields remain unchanged
// - Multiple partial updates work correctly
// - Array fields (languages) can be updated independently
func TestProjectPartialUpdate(t *testing.T) {
	adminClient := LoginAsAdmin(t)

	// Create a new project
	projectData := map[string]interface{}{
		"name":           "Original Project",
		"doc_url":        "https://example.com/original",
		"repo_url":       "https://github.com/original/docs",
		"languages":      []string{"en", "es"},
		"build_command":  "npm install",
		"export_command": "npm run export",
		"preview_command": "npm run preview",
	}

	resp := adminClient.MakeRequest(t, "POST", "/projects", projectData)
	adminClient.AssertCreated(t, resp)
	var project map[string]interface{}
	adminClient.ParseJSON(t, resp, &project)
	projectID := int(project["id"].(float64))

	// Update only the name
	resp = adminClient.MakeRequest(t, "PUT", fmt.Sprintf("/projects/%d", projectID), map[string]interface{}{
		"name": "Updated Name Only",
	})
	adminClient.AssertSuccess(t, resp)
	var updatedProject map[string]interface{}
	adminClient.ParseJSON(t, resp, &updatedProject)
	require.Equal(t, "Updated Name Only", updatedProject["name"])
	require.Equal(t, projectData["doc_url"], updatedProject["doc_url"]) // Should remain unchanged
	require.Equal(t, projectData["repo_url"], updatedProject["repo_url"])
	require.Equal(t, projectData["languages"], interfaceSliceToStringSlice(updatedProject["languages"].([]interface{})))
	require.Equal(t, projectData["build_command"], updatedProject["build_command"])
	require.Equal(t, projectData["export_command"], updatedProject["export_command"])
	require.Equal(t, projectData["preview_command"], updatedProject["preview_command"])

	// Update only languages
	resp = adminClient.MakeRequest(t, "PUT", fmt.Sprintf("/projects/%d", projectID), map[string]interface{}{
		"languages": []string{"en", "es", "fr", "de"},
	})
	adminClient.AssertSuccess(t, resp)
	adminClient.ParseJSON(t, resp, &updatedProject)
	require.Equal(t, "Updated Name Only", updatedProject["name"]) // Should remain unchanged
	require.Equal(t, []string{"en", "es", "fr", "de"}, interfaceSliceToStringSlice(updatedProject["languages"].([]interface{})))
	require.Equal(t, projectData["doc_url"], updatedProject["doc_url"])
	require.Equal(t, projectData["repo_url"], updatedProject["repo_url"])
	require.Equal(t, projectData["build_command"], updatedProject["build_command"])
	require.Equal(t, projectData["export_command"], updatedProject["export_command"])
	require.Equal(t, projectData["preview_command"], updatedProject["preview_command"])

	// Clean up
	resp = adminClient.MakeRequest(t, "DELETE", fmt.Sprintf("/projects/%d", projectID), nil)
	adminClient.AssertNoContent(t, resp)
}

// TestProjectValidation tests input validation and error handling for project endpoints.
// It validates that the API properly rejects invalid requests and returns appropriate errors:
// - Creating projects with missing required fields (name, doc_url, repo_url, build_command) returns 400 Bad Request
// - Creating projects with empty required fields returns 400 Bad Request
// - Accessing non-existent projects returns 404 Not Found
// - Updating non-existent projects returns 404 Not Found
// - Deleting non-existent projects returns 404 Not Found
func TestProjectValidation(t *testing.T) {
	adminClient := LoginAsAdmin(t)

	// Test creating project with missing required fields
	resp := adminClient.MakeRequest(t, "POST", "/projects", map[string]interface{}{
		"name": "Test Project",
		// Missing doc_url, repo_url, build_command
		"languages": []string{"en"},
	})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Test creating project with empty name
	resp = adminClient.MakeRequest(t, "POST", "/projects", map[string]interface{}{
		"name":          "",
		"doc_url":       "https://example.com/docs",
		"repo_url":      "https://github.com/test/docs",
		"build_command": "npm run build",
		"languages":     []string{"en"},
	})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Test creating project with empty doc_url
	resp = adminClient.MakeRequest(t, "POST", "/projects", map[string]interface{}{
		"name":          "Test Project",
		"doc_url":       "",
		"repo_url":      "https://github.com/test/docs",
		"build_command": "npm run build",
		"languages":     []string{"en"},
	})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Test creating project with empty repo_url
	resp = adminClient.MakeRequest(t, "POST", "/projects", map[string]interface{}{
		"name":          "Test Project",
		"doc_url":       "https://example.com/docs",
		"repo_url":      "",
		"build_command": "npm run build",
		"languages":     []string{"en"},
	})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Test creating project with empty build_command
	resp = adminClient.MakeRequest(t, "POST", "/projects", map[string]interface{}{
		"name":          "Test Project",
		"doc_url":       "https://example.com/docs",
		"repo_url":      "https://github.com/test/docs",
		"build_command": "",
		"languages":     []string{"en"},
	})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Test accessing non-existent project
	resp = adminClient.MakeRequest(t, "GET", "/projects/99999", nil)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	// Test updating non-existent project
	resp = adminClient.MakeRequest(t, "PUT", "/projects/99999", map[string]interface{}{
		"name": "Updated Name",
	})
	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	// Test deleting non-existent project
	resp = adminClient.MakeRequest(t, "DELETE", "/projects/99999", nil)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
