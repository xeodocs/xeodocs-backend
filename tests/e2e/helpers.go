package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

const gatewayURL = "http://localhost:8080/v1"

type TestClient struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

func NewTestClient(baseURL string) *TestClient {
	return &TestClient{
		httpClient: &http.Client{},
		baseURL:    baseURL,
	}
}

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

func (c *TestClient) AssertSuccess(t *testing.T, resp *http.Response) {
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func (c *TestClient) AssertCreated(t *testing.T, resp *http.Response) {
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func (c *TestClient) AssertNoContent(t *testing.T, resp *http.Response) {
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func (c *TestClient) ParseJSON(t *testing.T, resp *http.Response, v interface{}) {
	defer resp.Body.Close()
	err := json.NewDecoder(resp.Body).Decode(v)
	require.NoError(t, err)
}

func LoginAsAdmin(t *testing.T) *TestClient {
	client := NewTestClient(gatewayURL)
	client.Login(t, "admin", "tempadmin123")
	return client
}
