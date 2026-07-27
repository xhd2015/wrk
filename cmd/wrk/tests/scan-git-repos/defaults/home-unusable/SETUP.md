# Scenario

**Feature**: bare `wrk --scan-git-repos` fails clearly when `$HOME` is not a directory

```
# HOME points at a regular file under WorkRoot (not a directory)
HOME=WorkRoot/not-a-dir-home  (file)
  -> wrk --scan-git-repos
  -> non-zero exit
  -> stderr explains unusable home / ~ (must not require ~/Projects)
```

## Preconditions

- `FakeHome` is a **file** under WorkRoot (not a directory). Using a missing path is
  unreliable: some tools auto-create `$HOME` (and CI/macOS may race mkdir before
  Stat), so a file is a stable "not a directory" fixture.
- No scan fixtures needed — default-root resolution fails before discovery.

## Steps

1. Create `{WorkRoot}/not-a-dir-home` as a regular file; set `FakeHome` to it.
2. Set Args to bare `--scan-git-repos`.

```go
import (
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	// File-as-home: stable rejection (missing dirs can be auto-created by tooling).
	homeFile := filepath.Join(req.WorkRoot, "not-a-dir-home")
	if err := os.WriteFile(homeFile, []byte("not a directory\n"), 0o644); err != nil {
		t.Fatalf("write fake home file: %v", err)
	}
	req.FakeHome = homeFile
	req.Args = []string{"--scan-git-repos"}
	return nil
}
```
