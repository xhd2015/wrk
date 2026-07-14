# Scenario

**Feature**: wrk --list without -v has empty stderr

```
main repo -> wrk --list -> stderr empty
```

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	repo := filepath.Join(req.WorkRoot, "off-list-main")
	initFetchVerboseRepo(t, repo, "off list main")
	req.RepoDir = repo
	req.Args = []string{"--list"}
	return nil
}
```