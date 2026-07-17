# Scenario

**Feature**: `--json` is only valid with bare `--tag-next`; multi-stage + `--json` is rejected

```
myrepo -> wrk --sync --tag-next --json
  -> non-zero
  -> stderr names --json and rejects multi-stage combination
```

## Steps

1. Minimal main repo.
2. Run multi-stage with `--json`.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.Args = []string{"--sync", "--tag-next", "--json"}
	return nil
}
```
