package repository

type CloneRepoRequest struct {
	RepoURL  string `json:"repoUrl"`
	ProjectID int    `json:"projectId"`
}

type CreateLanguageCopiesRequest struct {
	ProjectID int      `json:"projectId"`
	Languages []string `json:"languages"`
}

type SyncRepoRequest struct {
	ProjectID int `json:"projectId"`
}

type DeleteRepoRequest struct {
	ProjectID int `json:"projectId"`
}

type RepoResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
