# Scenario

**Feature**: --task with empty description produces an error

```
wrk --task "" -> non-zero exit, stderr says task description must not be empty
```

## Steps

1. Create repo.
2. Use req.Args = ["--task", ""] to pass the empty value.
3. Verify non-zero exit and error message.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	req.RepoDir = mainRepo
	req.Args = []string{"--task", ""}
	return nil
}
```