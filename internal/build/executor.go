package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/xeodocs/xeodocs-backend/internal/project"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
)

// ExecuteBuild executes the build command for a project
func ExecuteBuild(projectID int, cfg *config.Config) error {
	proj, err := project.GetProjectByID(projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	if proj.BuildCommand == "" {
		return fmt.Errorf("no build command configured for project")
	}

	// Assume repos are stored in /repos/{projectID}
	repoPath := filepath.Join("/repos", strconv.Itoa(projectID))

	// Check if repo directory exists
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return fmt.Errorf("repository directory does not exist: %s", repoPath)
	}

	// Execute build command
	return executeCommand(proj.BuildCommand, repoPath)
}

// ExecuteExport executes the export command for a project
func ExecuteExport(projectID int, cfg *config.Config) error {
	proj, err := project.GetProjectByID(projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	if proj.ExportCommand == "" {
		return fmt.Errorf("no export command configured for project")
	}

	// Assume repos are stored in /repos/{projectID}
	repoPath := filepath.Join("/repos", strconv.Itoa(projectID))

	// Check if repo directory exists
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return fmt.Errorf("repository directory does not exist: %s", repoPath)
	}

	// Execute export command
	return executeCommand(proj.ExportCommand, repoPath)
}

// ExecutePreview executes the preview command for a project
func ExecutePreview(projectID int, cfg *config.Config) error {
	proj, err := project.GetProjectByID(projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	if proj.PreviewCommand == "" {
		return fmt.Errorf("no preview command configured for project")
	}

	// Assume repos are stored in /repos/{projectID}
	repoPath := filepath.Join("/repos", strconv.Itoa(projectID))

	// Check if repo directory exists
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return fmt.Errorf("repository directory does not exist: %s", repoPath)
	}

	// Execute preview command (this might run in background for servers)
	return executeCommand(proj.PreviewCommand, repoPath)
}

// executeCommand executes a shell command in the specified directory
func executeCommand(command, dir string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}
