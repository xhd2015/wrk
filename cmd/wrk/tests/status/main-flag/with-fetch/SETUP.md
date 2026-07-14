# Scenario

**Feature**: --fetch is allowed with --main --status (not treated as exclusive)

```
wrk --main --status --fetch -> exit 0; status of main (may show (no upstream))
```

## Steps

1. Create clean main repo (no upstream).
2. cwd = main; Args = `--main`, `--status`, `--fetch`.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	statusInitRepoWithSubject(t, mainRepo, "main status with fetch")
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	setMainStatusArgs(req, "--main", "--status", "--fetch")
	return nil
}
```