package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Service struct {
	token string
	owner string
}

func NewService(token, owner string) *Service {
	return &Service{
		token: token,
		owner: owner,
	}
}

// EnsureForkExists checks if a fork with the target name exists in the configured owner's account.
// If it doesn't exist, it attempts to fork the source repo to the target name.
// sourceRepoURL should be the full URL (e.g., https://github.com/original/repo)
// targetName is the name for the fork (e.g., project-slug)
func (s *Service) EnsureForkExists(sourceRepoURL, targetName string) (string, error) {
	// 1. Parse source owner and repo from URL
	sourceOwner, sourceRepo, err := parseRepoURL(sourceRepoURL)
	if err != nil {
		return "", fmt.Errorf("invalid source repo URL: %w", err)
	}

	// 2. Check if the fork already exists
	exists, err := s.checkRepoExists(s.owner, targetName)
	if err != nil {
		return "", fmt.Errorf("failed to check if repo exists: %w", err)
	}

	if exists {
		return fmt.Sprintf("https://github.com/%s/%s", s.owner, targetName), nil
	}

	// 3. Create the fork
	if err := s.createFork(sourceOwner, sourceRepo, targetName); err != nil {
		return "", fmt.Errorf("failed to create fork: %w", err)
	}

	return fmt.Sprintf("https://github.com/%s/%s", s.owner, targetName), nil
}

func parseRepoURL(url string) (owner, repo string, err error) {
	// Remove trailing slash
	url = strings.TrimSuffix(url, "/")

	// Assuming format https://github.com/owner/repo or similar
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("url too short")
	}

	repo = parts[len(parts)-1]
	owner = parts[len(parts)-2]

	// Remove .git if present
	repo = strings.TrimSuffix(repo, ".git")

	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("could not parse owner/repo")
	}

	return owner, repo, nil
}

func (s *Service) checkRepoExists(owner, repo string) (bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	s.addHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}

func (s *Service) createFork(sourceOwner, sourceRepo, targetName string) error {
	// https://docs.github.com/en/rest/repos/forks?apiVersion=2022-11-28#create-a-fork
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/forks", sourceOwner, sourceRepo)

	body := map[string]interface{}{
		"name":                targetName,
		"default_branch_only": true,
	}

	// Check if the target owner is an organization
	ownerType, err := s.getOwnerType(s.owner)
	if err != nil {
		return fmt.Errorf("failed to determine owner type: %w", err)
	}

	if ownerType == "Organization" {
		body["organization"] = s.owner
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	s.addHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusOK {
		return nil
	}

	// Read body for error details
	var errResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
		if msg, ok := errResp["message"].(string); ok {
			return fmt.Errorf("failed to create fork (status %d): %s", resp.StatusCode, msg)
		}
	}

	return fmt.Errorf("failed to create fork, status: %d", resp.StatusCode)
}

func (s *Service) getOwnerType(owner string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s", owner)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	s.addHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get owner info, status: %d", resp.StatusCode)
	}

	var userResp struct {
		Type string `json:"type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return "", err
	}

	return userResp.Type, nil
}

func (s *Service) addHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
}
