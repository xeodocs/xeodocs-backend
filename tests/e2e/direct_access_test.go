package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDirectServiceAccess demonstrates how to test internal service endpoints directly
// This test shows how to access services without going through the gateway
func TestDirectServiceAccess(t *testing.T) {
	// Create clients for direct service access
	authClient := NewAuthClient()
	projectClient := NewProjectClient()

	// Example: Test direct auth service health check (if endpoint exists)
	// Note: This assumes the auth service has a health endpoint
	resp := authClient.MakeRequest(t, "GET", "/health", nil)
	// This would fail if the endpoint doesn't exist, but shows the pattern

	// For demonstration, let's test a simple request
	// In a real scenario, you'd test actual internal endpoints

	// Example with project service
	resp = projectClient.MakeRequest(t, "GET", "/health", nil)
	// Again, this assumes a health endpoint exists

	// Close responses
	if resp != nil {
		resp.Body.Close()
	}
}

// TestGatewayVsDirectAccess shows the difference between gateway and direct access
func TestGatewayVsDirectAccess(t *testing.T) {
	// Gateway client (external access through gateway)
	gatewayClient := NewGatewayClient()

	// Direct auth service client (internal access)
	authClient := NewAuthClient()

	// Login through gateway
	adminClient := LoginAsAdmin(t)

	// Now you could compare responses from gateway vs direct access
	// For example, testing auth endpoints:

	// Through gateway (current way)
	resp1 := adminClient.MakeRequest(t, "POST", "/auth/register", map[string]interface{}{
		"username": "gateway-user",
		"password": "password123",
		"role":     "viewer",
	})
	if resp1.StatusCode == http.StatusOK || resp1.StatusCode == http.StatusCreated {
		resp1.Body.Close()
	}

	// Direct access (new way) - assuming auth service has direct endpoints
	// Note: This would depend on what endpoints the auth service exposes internally
	resp2 := authClient.MakeRequest(t, "GET", "/", nil)
	if resp2.StatusCode != 0 { // Check if request was made
		resp2.Body.Close()
	}

	// Use gatewayClient in a simple request to demonstrate
	resp3 := gatewayClient.MakeRequest(t, "GET", "/health", nil)
	require.Equal(t, http.StatusOK, resp3.StatusCode)
	resp3.Body.Close()
}
