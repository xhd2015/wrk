# Scenario

**Feature**: bare `wrk --scan-git-repos` fails clearly when `$HOME` is missing or not a directory

```
# HOME points at a non-existent path under WorkRoot
HOME=WorkRoot/missing-home
  -> wrk --scan-git-repos
  -> non-zero exit
  -> stderr explains unusable home / ~ (must not require ~/Projects)
```

## Preconditions

- `FakeHome` is a path under WorkRoot that is **never created** (not a directory).
- No scan fixtures needed — default-root resolution fails before discovery.

## Steps

1. Set `FakeHome` to `{WorkRoot}/missing-home` without creating it.
2. Set Args to bare `--scan-git-repos`.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	// Non-existent home: UserHomeDir returns the path; product must reject it.
	req.FakeHome = filepath.Join(req.WorkRoot, "missing-home")
	req.Args = []string{"--scan-git-repos"}
	return nil
}
```
