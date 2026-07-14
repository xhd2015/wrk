# Scenario

**Feature**: --main --status remains mutually exclusive with other standalone modes

```
wrk --main --status --list -> non-zero; mutually exclusive
```

## Preconditions

- Git checkout exists so rejection is mode composition, not missing git.

## Steps

- Descendants combine `--main`, `--status`, and another mode flag.

## Context

- Pure `wrk --main --list` is already covered under `main/errors/mutual-exclusion/with-list`
  and is not re-tested here.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	repo := filepath.Join(req.WorkRoot, "myrepo")
	statusInitRepoWithSubject(t, repo, "main status exclusive")
	req.MainRepo = repo
	req.RepoDir = repo
	return nil
}
```