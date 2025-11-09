package repository

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
)

func CloneRepoHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req CloneRepoRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Clone repo to /repos/projectID
		repoPath := fmt.Sprintf("/repos/%d", req.ProjectID)
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			log.Printf("Error creating repo directory: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		_, err := git.PlainClone(repoPath, false, &git.CloneOptions{
			URL: req.RepoURL,
		})
		if err != nil {
			log.Printf("Error cloning repo: %v", err)
			http.Error(w, "Failed to clone repository", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RepoResponse{Success: true, Message: "Repository cloned successfully"})
	}
}

func CreateLanguageCopiesHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req CreateLanguageCopiesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		repoPath := fmt.Sprintf("/repos/%d", req.ProjectID)
		for _, lang := range req.Languages {
			langPath := fmt.Sprintf("/repos/%d/%s", req.ProjectID, lang)
			if err := copyDir(repoPath, langPath); err != nil {
				log.Printf("Error copying to language dir %s: %v", lang, err)
				http.Error(w, "Failed to create language copies", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RepoResponse{Success: true, Message: "Language copies created successfully"})
	}
}

func SyncRepoHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req SyncRepoRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		repoPath := fmt.Sprintf("/repos/%d", req.ProjectID)
		repo, err := git.PlainOpen(repoPath)
		if err != nil {
			log.Printf("Error opening repo: %v", err)
			http.Error(w, "Repository not found", http.StatusNotFound)
			return
		}

		worktree, err := repo.Worktree()
		if err != nil {
			log.Printf("Error getting worktree: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		err = worktree.Pull(&git.PullOptions{})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			log.Printf("Error pulling repo: %v", err)
			http.Error(w, "Failed to sync repository", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RepoResponse{Success: true, Message: "Repository synced successfully"})
	}
}

func DeleteRepoHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req DeleteRepoRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		repoPath := fmt.Sprintf("/repos/%d", req.ProjectID)
		if err := os.RemoveAll(repoPath); err != nil {
			log.Printf("Error deleting repo: %v", err)
			http.Error(w, "Failed to delete repository", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// Helper function to copy directory
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, data, info.Mode())
	})
}
