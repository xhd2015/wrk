package wrkserver

// ListProjectsResponse is the GET {base}/projects envelope.
type ListProjectsResponse struct {
	Projects []ProjectStatus `json:"projects"`
}

// ProjectStatus is one registered main repository and its linked worktrees.
type ProjectStatus struct {
	Path      string           `json:"path"`
	Name      string           `json:"name"`
	Branch    string           `json:"branch,omitempty"`
	Commit    string           `json:"commit,omitempty"`
	Subject   string           `json:"subject,omitempty"`
	Clean     bool             `json:"clean"`
	Error     string           `json:"error,omitempty"`
	Worktrees []WorktreeStatus `json:"worktrees"`
}

// WorktreeStatus is one linked worktree (main is not listed here).
type WorktreeStatus struct {
	Path   string `json:"path"`
	Name   string `json:"name"`
	Branch string `json:"branch,omitempty"`
	Clean  bool   `json:"clean"`
	IsMain bool   `json:"is_main"`
	Error  string `json:"error,omitempty"`
}

// CreateWorktreeRequest is the POST {base}/worktrees body.
type CreateWorktreeRequest struct {
	ProjectPath string `json:"project_path"`
	Task        string `json:"task,omitempty"`
}

// CreateWorktreeResponse is the successful create payload.
type CreateWorktreeResponse struct {
	Path   string `json:"path"`
	Branch string `json:"branch"`
}

// ErrorBody is the JSON error envelope for 4xx/5xx.
type ErrorBody struct {
	Error string `json:"error"`
}
