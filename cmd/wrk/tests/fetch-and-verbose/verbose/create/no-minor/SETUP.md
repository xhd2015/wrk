# Scenario

**Feature**: wrk -v create does not log minor git reads

```
stderr has no rev-parse or status log lines
```

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	repo := filepath.Join(req.WorkRoot, "create-v-nominor")
	initFetchVerboseRepo(t, repo, "create v nominor")
	req.RepoDir = repo
	req.Args = []string{"-v"}
	return nil
}
```