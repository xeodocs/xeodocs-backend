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

// GetLatestCommit fetches the SHA of the latest commit on the specified branch of the repo.
// repoURL should be the full URL.
func (s *Service) GetLatestCommit(repoURL, branch string) (string, error) {
	owner, repo, err := parseRepoURL(repoURL)
	if err != nil {
		return "", err
	}

	if branch == "" {
		branch = "main" // Default to main if not specified, though ideally should be passed
	}

	// https://docs.github.com/en/rest/commits/commits?apiVersion=2022-11-28#list-commits
	// We just want the latest one
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?sha=%s&per_page=1", owner, repo, branch)

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
		return "", fmt.Errorf("failed to list commits, status: %d", resp.StatusCode)
	}

	var commits []struct {
		Sha string `json:"sha"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return "", err
	}

	if len(commits) == 0 {
		return "", fmt.Errorf("no commits found")
	}

	return commits[0].Sha, nil
}

// SyncFork triggers a sync of the fork with the upstream repository.
// forkName is the name of the fork (e.g. project slug).
// branch is the branch to sync.
func (s *Service) SyncFork(forkName, branch string) error {
	// https://docs.github.com/en/rest/branches/branches?apiVersion=2022-11-28#sync-a-fork-branch-with-the-upstream-repository
	// Endpoint: POST /repos/{owner}/{repo}/merge-upstream

	if branch == "" {
		branch = "main"
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/merge-upstream", s.owner, forkName)

	body := map[string]string{
		"branch": branch,
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

	// Status: 200 OK (successfully merged), 409 Conflict (merge conflict), 422 Unprocessable Entity (branch not found or other issues)
	// If the branch is already up to date, it might return 200 with specific message, or maybe 409?
	// Docs say:
	// 200: The branch has been successfully synced
	// 409: Merge conflict
	// 422: The branch could not be synced (e.g. because it's not a fork?)

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	var errResp map[string]interface{}
	// Try to read error message
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
		if msg, ok := errResp["message"].(string); ok {
			// If message says "This branch is not behind the upstream", we consider it success?
			// The API returns 200 even if it's up to date?
			// Docs say "This branch is not behind the upstream" is a possible message for 422? No, usually 200 means success or no-op?
			// Actually, if it is up to date, it might return 422 with "This branch is not behind the upstream <branch>".
			// Let's handle that case gracefully.
			if resp.StatusCode == http.StatusUnprocessableEntity && strings.Contains(msg, "not behind the upstream") {
				return nil
			}
			return fmt.Errorf("failed to sync fork (status %d): %s", resp.StatusCode, msg)
		}
	}

	return fmt.Errorf("failed to sync fork, status: %d", resp.StatusCode)
}

// EnsureBranchExists verifies if a branch exists in the specified repo.
// If it does not exist, it creates it from the baseBranch.
// repoName is the name of the repository (e.g. fork name).
func (s *Service) EnsureBranchExists(repoName, targetBranch, baseBranch string) error {
	exists, err := s.checkBranchExists(repoName, targetBranch)
	if err != nil {
		return fmt.Errorf("failed to check branch existence: %w", err)
	}
	if exists {
		return nil
	}

	// Need the SHA of the base branch to create the new one
	baseSha, err := s.GetLatestCommit(fmt.Sprintf("https://github.com/%s/%s", s.owner, repoName), baseBranch)
	if err != nil {
		return fmt.Errorf("failed to get base branch sha: %w", err)
	}

	if err := s.createBranch(repoName, targetBranch, baseSha); err != nil {
		return fmt.Errorf("failed to create branch %s: %w", targetBranch, err)
	}

	return nil
}

func (s *Service) checkBranchExists(repoName, branch string) (bool, error) {
	// GET /repos/{owner}/{repo}/branches/{branch}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches/%s", s.owner, repoName, branch)
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

	return false, fmt.Errorf("unexpected status checking branch: %d", resp.StatusCode)
}

func (s *Service) createBranch(repoName, branch, sha string) error {
	// POST /repos/{owner}/{repo}/git/refs
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs", s.owner, repoName)

	body := map[string]string{
		"ref": fmt.Sprintf("refs/heads/%s", branch),
		"sha": sha,
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

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		return nil
	}

	var errResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
		if msg, ok := errResp["message"].(string); ok {
			return fmt.Errorf("api error: %s", msg)
		}
	}

	return fmt.Errorf("unexpected status creating branch: %d", resp.StatusCode)
}

func (s *Service) addHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
}
