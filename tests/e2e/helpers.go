package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const gatewayURL = "http://localhost:12020/v1"

// Internal service URLs for testing (when running from within Docker network)
var (
	authServiceURL       = getEnvOrDefault("AUTH_SERVICE_URL", "http://auth:80")
	projectServiceURL    = getEnvOrDefault("PROJECT_SERVICE_URL", "http://project:80")
	loggingServiceURL    = getEnvOrDefault("LOGGING_SERVICE_URL", "http://logging:80")
	buildServiceURL      = getEnvOrDefault("BUILD_SERVICE_URL", "http://build:80")
	analyticsServiceURL  = getEnvOrDefault("ANALYTICS_SERVICE_URL", "http://analytics:80")
	repositoryServiceURL = getEnvOrDefault("REPOSITORY_SERVICE_URL", "http://repository:80")
)

// getEnvOrDefault returns the value of the environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestClient is a helper struct for making authenticated HTTP requests
// to the API gateway during end-to-end tests. It manages authentication
// tokens and provides convenient methods for common HTTP operations.
type TestClient struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

// NewTestClient creates a new TestClient instance with the specified base URL.
// The client is initialized with a default HTTP client and no authentication token.
func NewTestClient(baseURL string) *TestClient {
	return &TestClient{
		httpClient: &http.Client{},
		baseURL:    baseURL,
	}
}

// Login authenticates the TestClient with the API using the provided username and password.
// It makes a POST request to /auth/login, expects a 200 OK response with a JWT token,
// and stores the token for use in subsequent authenticated requests.
func (c *TestClient) Login(t *testing.T, username, password string) {
	reqBody := map[string]string{"username": username, "password": password}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	resp, err := c.httpClient.Post(c.baseURL+"/auth/login", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var res map[string]string
	err = json.NewDecoder(resp.Body).Decode(&res)
	require.NoError(t, err)
	require.NotEmpty(t, res["token"])

	c.token = res["token"]
}

// MakeRequest performs an HTTP request to the API with optional JSON body and authentication.
// It automatically sets Content-Type: application/json for requests with a body,
// and Authorization: Bearer <token> header if the client has been authenticated.
// Returns the raw HTTP response for further inspection and assertion.
func (c *TestClient) MakeRequest(t *testing.T, method, path string, body interface{}) *http.Response {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	require.NoError(t, err)

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	require.NoError(t, err)

	return resp
}

// AssertSuccess asserts that the HTTP response has a 200 OK status code.
// This is used for successful GET, PUT, and other operations that return data.
func (c *TestClient) AssertSuccess(t *testing.T, resp *http.Response) {
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// AssertCreated asserts that the HTTP response has a 201 Created status code.
// This is used for successful POST operations that create new resources.
func (c *TestClient) AssertCreated(t *testing.T, resp *http.Response) {
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

// AssertNoContent asserts that the HTTP response has a 204 No Content status code.
// This is used for successful DELETE operations and other operations that return no data.
func (c *TestClient) AssertNoContent(t *testing.T, resp *http.Response) {
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// ParseJSON decodes the JSON response body into the provided interface.
// It closes the response body automatically and fails the test if JSON parsing fails.
func (c *TestClient) ParseJSON(t *testing.T, resp *http.Response, v interface{}) {
	defer resp.Body.Close()
	err := json.NewDecoder(resp.Body).Decode(v)
	require.NoError(t, err)
}

// LoginAsAdmin creates a new TestClient and authenticates it as the default admin user.
// This is a convenience function for tests that need admin-level access to the API.
// Uses the hardcoded admin credentials: username="admin", password="tempadmin123".
func LoginAsAdmin(t *testing.T) *TestClient {
	client := NewTestClient(gatewayURL)
	client.Login(t, "admin", "tempadmin123")
	return client
}

// NewAuthClient creates a TestClient for direct access to the auth service
func NewAuthClient() *TestClient {
	return NewTestClient(authServiceURL)
}

// NewProjectClient creates a TestClient for direct access to the project service
func NewProjectClient() *TestClient {
	return NewTestClient(projectServiceURL)
}

// NewLoggingClient creates a TestClient for direct access to the logging service
func NewLoggingClient() *TestClient {
	return NewTestClient(loggingServiceURL)
}

// NewBuildClient creates a TestClient for direct access to the build service
func NewBuildClient() *TestClient {
	return NewTestClient(buildServiceURL)
}

// NewAnalyticsClient creates a TestClient for direct access to the analytics service
func NewAnalyticsClient() *TestClient {
	return NewTestClient(analyticsServiceURL)
}

// NewRepositoryClient creates a TestClient for direct access to the repository service
func NewRepositoryClient() *TestClient {
	return NewTestClient(repositoryServiceURL)
}

// NewGatewayClient creates a TestClient for access through the API gateway
func NewGatewayClient() *TestClient {
	return NewTestClient(gatewayURL)
}
