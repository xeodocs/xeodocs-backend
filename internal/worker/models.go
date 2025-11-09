package worker

// Task represents a message from the queue
type Task struct {
	Type    string                 `json:"type"`    // e.g., "clone_repo", "translate_files", "build_task"
	Payload map[string]interface{} `json:"payload"`
	ID      string                 `json:"id"`
}
