# Scenario

**Feature**: wrk --rm accepts stale path after repo deleted

```
wrk --add myrepo -> delete .git -> wrk --rm <old-path> -> entry removed, stdout main path
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo`.
2. Run `wrk --add <mainRepo>` to record.
3. Delete `{WorkRoot}/myrepo/.git` (directory remains).
4. Run `wrk --rm <mainRepo>` using the original absolute path.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := initProjectsRepo(t, req.WorkRoot, "myrepo")
	recordedPath := resolvePath(t, mainRepo)
	recordProjectViaAdd(t, req, mainRepo)
	removeGitDir(t, mainRepo)
	req.MainRepo = recordedPath
	req.Args = []string{"--rm", recordedPath}
	return nil
}
```
