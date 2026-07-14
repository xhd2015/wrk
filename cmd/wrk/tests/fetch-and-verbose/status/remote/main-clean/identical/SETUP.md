# Scenario

**Feature**: main checkout with upstream shows Remote: identical

```
tracked repo up-to-date + optional linked wt -> root Remote: identical; linked has Master: only
```

```go
func Setup(t *testing.T, req *Request) error {
	origin := setupFetchVerboseBareOrigin(t, req.WorkRoot, "origin")
	repo := setupFetchVerboseTrackedRepo(t, req.WorkRoot, "status-ident", origin, "status ident base")
	wtDir := addLinkedWorktreeInRepo(t, repo, "wt-linked", "wt-ident")
	req.MainRepo = repo
	req.WtDir = wtDir
	req.RepoDir = repo
	req.Args = []string{"--status"}
	return nil
}
```