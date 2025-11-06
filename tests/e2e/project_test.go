package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func interfaceSliceToStringSlice(slice []interface{}) []string {
	result := make([]string, len(slice))
	for i, v := range slice {
		result[i] = v.(string)
	}
	return result
}

func TestProjectCRUD(t *testing.T) {
	adminClient := LoginAsAdmin(t)

	// Create a new project
	projectData := map[string]interface{}{
		"name":          "Test Documentation Project",
		"repo_url":      "https://github.com/example/docs",
		"languages":     []string{"en", "es", "fr"},
		"build_commands": []string{"npm install", "npm run build"},
	}

	resp := adminClient.MakeRequest(t, "POST", "/projects", projectData)
	adminClient.AssertCreated(t, resp)
	var createdProject map[string]interface{}
	adminClient.ParseJSON(t, resp, &createdProject)
	require.NotEmpty(t, createdProject["id"])
	require.Equal(t, projectData["name"], createdProject["name"])
	require.Equal(t, projectData["repo_url"], createdProject["repo_url"])
	require.Equal(t, projectData["languages"], interfaceSliceToStringSlice(createdProject["languages"].([]interface{})))
	require.Equal(t, projectData["build_commands"], interfaceSliceToStringSlice(createdProject["build_commands"].([]interface{})))
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
	require.Equal(t, projectData["repo_url"], fetchedProject["repo_url"])
	require.Equal(t, projectData["languages"], interfaceSliceToStringSlice(fetchedProject["languages"].([]interface{})))
	require.Equal(t, projectData["build_commands"], interfaceSliceToStringSlice(fetchedProject["build_commands"].([]interface{})))

	// Update the project
	updatedData := map[string]interface{}{
		"name":          "Updated Documentation Project",
		"repo_url":      "https://github.com/example/updated-docs",
		"languages":     []string{"en", "es", "fr", "de"},
		"build_commands": []string{"npm install", "npm run build", "npm run deploy"},
	}

	resp = adminClient.MakeRequest(t, "PUT", fmt.Sprintf("/projects/%d", projectID), updatedData)
	adminClient.AssertSuccess(t, resp)
	var updatedProject map[string]interface{}
	adminClient.ParseJSON(t, resp, &updatedProject)
	require.Equal(t, projectID, int(updatedProject["id"].(float64)))
	require.Equal(t, updatedData["name"], updatedProject["name"])
	require.Equal(t, updatedData["repo_url"], updatedProject["repo_url"])
	require.Equal(t, updatedData["languages"], interfaceSliceToStringSlice(updatedProject["languages"].([]interface{})))
	require.Equal(t, updatedData["build_commands"], interfaceSliceToStringSlice(updatedProject["build_commands"].([]interface{})))
	require.NotEqual(t, createdProject["updated_at"], updatedProject["updated_at"])

	// Delete the project
	resp = adminClient.MakeRequest(t, "DELETE", fmt.Sprintf("/projects/%d", projectID), nil)
	adminClient.AssertNoContent(t, resp)

	// Verify deletion - should get 404
	resp = adminClient.MakeRequest(t, "GET", fmt.Sprintf("/projects/%d", projectID), nil)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestProjectUserIsolation(t *testing.T) {
	adminClient := LoginAsAdmin(t)

	// Create a project as admin
	projectData := map[string]interface{}{
		"name":          "Admin Project",
		"repo_url":      "https://github.com/admin/docs",
		"languages":     []string{"en"},
		"build_commands": []string{"npm install"},
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

	// Try to access admin's project - should fail
	resp = newUserClient.MakeRequest(t, "GET", fmt.Sprintf("/projects/%d", adminProjectID), nil)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	// Create a project as the new user
	userProjectData := map[string]interface{}{
		"name":          "User Project",
		"repo_url":      "https://github.com/user/docs",
		"languages":     []string{"en", "fr"},
		"build_commands": []string{"yarn install", "yarn build"},
	}

	resp = newUserClient.MakeRequest(t, "POST", "/projects", userProjectData)
	newUserClient.AssertCreated(t, resp)
	var userProject map[string]interface{}
	newUserClient.ParseJSON(t, resp, &userProject)
	userProjectID := int(userProject["id"].(float64))

	// User should be able to access their own project
	resp = newUserClient.MakeRequest(t, "GET", fmt.Sprintf("/projects/%d", userProjectID), nil)
	newUserClient.AssertSuccess(t, resp)
	var fetchedUserProject map[string]interface{}
	newUserClient.ParseJSON(t, resp, &fetchedUserProject)
	require.Equal(t, userProjectData["name"], fetchedUserProject["name"])

	// Admin should not be able to access user's project
	resp = adminClient.MakeRequest(t, "GET", fmt.Sprintf("/projects/%d", userProjectID), nil)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	// Admin should only see their own projects in list
	resp = adminClient.MakeRequest(t, "GET", "/projects", nil)
	adminClient.AssertSuccess(t, resp)
	var adminProjects []map[string]interface{}
	adminClient.ParseJSON(t, resp, &adminProjects)

	// Should only see admin's project, not user's project
	adminProjectFound := false
	for _, p := range adminProjects {
		if int(p["id"].(float64)) == adminProjectID {
			adminProjectFound = true
		}
		if int(p["id"].(float64)) == userProjectID {
			t.Errorf("Admin should not see user's project in their list")
		}
	}
	require.True(t, adminProjectFound, "Admin should see their own project")

	// User should only see their own projects in list
	resp = newUserClient.MakeRequest(t, "GET", "/projects", nil)
	newUserClient.AssertSuccess(t, resp)
	var userProjects []map[string]interface{}
	newUserClient.ParseJSON(t, resp, &userProjects)

	// Should only see user's project, not admin's project
	userProjectFound := false
	for _, p := range userProjects {
		if int(p["id"].(float64)) == userProjectID {
			userProjectFound = true
		}
		if int(p["id"].(float64)) == adminProjectID {
			t.Errorf("User should not see admin's project in their list")
		}
	}
	require.True(t, userProjectFound, "User should see their own project")

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

func TestProjectPartialUpdate(t *testing.T) {
	adminClient := LoginAsAdmin(t)

	// Create a new project
	projectData := map[string]interface{}{
		"name":          "Original Project",
		"repo_url":      "https://github.com/original/docs",
		"languages":     []string{"en", "es"},
		"build_commands": []string{"npm install", "npm run build"},
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
	require.Equal(t, projectData["repo_url"], updatedProject["repo_url"]) // Should remain unchanged
	require.Equal(t, projectData["languages"], interfaceSliceToStringSlice(updatedProject["languages"].([]interface{})))
	require.Equal(t, projectData["build_commands"], interfaceSliceToStringSlice(updatedProject["build_commands"].([]interface{})))

	// Update only languages
	resp = adminClient.MakeRequest(t, "PUT", fmt.Sprintf("/projects/%d", projectID), map[string]interface{}{
		"languages": []string{"en", "es", "fr", "de"},
	})
	adminClient.AssertSuccess(t, resp)
	adminClient.ParseJSON(t, resp, &updatedProject)
	require.Equal(t, "Updated Name Only", updatedProject["name"]) // Should remain unchanged
	require.Equal(t, []string{"en", "es", "fr", "de"}, interfaceSliceToStringSlice(updatedProject["languages"].([]interface{})))
	require.Equal(t, projectData["repo_url"], updatedProject["repo_url"])
	require.Equal(t, projectData["build_commands"], interfaceSliceToStringSlice(updatedProject["build_commands"].([]interface{})))

	// Clean up
	resp = adminClient.MakeRequest(t, "DELETE", fmt.Sprintf("/projects/%d", projectID), nil)
	adminClient.AssertNoContent(t, resp)
}

func TestProjectValidation(t *testing.T) {
	adminClient := LoginAsAdmin(t)

	// Test creating project with missing required fields
	resp := adminClient.MakeRequest(t, "POST", "/projects", map[string]interface{}{
		"name": "Test Project",
		// Missing repo_url
		"languages": []string{"en"},
	})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Test creating project with empty name
	resp = adminClient.MakeRequest(t, "POST", "/projects", map[string]interface{}{
		"name":     "",
		"repo_url": "https://github.com/test/docs",
		"languages": []string{"en"},
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
