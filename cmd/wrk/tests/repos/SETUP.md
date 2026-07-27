# Scenario

**Feature**: wrk --repos lists discovered repository paths

```
# cwd resolves to an effective git toplevel; repos mode scans that root
wrk --repos from cwd -> scan_repo-based repo paths

# repos is standalone; combining with another mode is rejected
wrk --repos + other mode -> error (mutually exclusive)
```

## Preconditions

- Git must be available.
- `wrk --repos` is a standalone mode.

## Steps

- Tests invoke `wrk --repos` by default with `req.Args = []string{"--repos"}`.
- Descendant scenarios choose whether cwd is inside a git checkout and whether another mode is also present.

## Context

- Successful repos output is one slash-normalized path per line.
- The root checkout is printed as `.`.
- Ordering matches the repository discovery ordering used by `wrk --status`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	req.Args = []string{"--repos"}
	return nil
}

func reposInitRepo(t *testing.T, path string) {
	t.Helper()
	mkdirAll(t, path)
	runGitIsolated(t, path, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, path, "config", "user.email", "test@test.com")
	runGitIsolated(t, path, "config", "user.name", "Test")
	writeFile(t, filepath.Join(path, "README.md"), "# "+filepath.Base(path)+"\n")
	runGitIsolated(t, path, "add", "README.md")
	runGitIsolated(t, path, "commit", "-m", "init "+filepath.Base(path))
}
```
