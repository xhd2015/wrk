# Scenario

**Feature**: successful bare --propagate-tags records events.jsonl command "propagate-tags"

```
# bare propagate-tags success -> events.jsonl last event command=propagate-tags
cwd=lib (tagged) -> wrk --propagate-tags --dry-run -> event appended
```

## Steps

1. Create tagged source module `example.com/lib` at `v1.0.0`.
2. Run `wrk --propagate-tags --dry-run` from the source (no consumers required).

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	libPath := filepath.Join(req.WorkRoot, "repos", "lib")
	initSingleModuleRepo(t, libPath, "example.com/lib", nil)
	tagRepo(t, libPath, "v1.0.0")
	libPath = resolvePath(t, libPath)

	req.SourcePath = libPath
	req.RepoDir = libPath
	req.Args = []string{"--propagate-tags", "--dry-run"}
	return nil
}
```
