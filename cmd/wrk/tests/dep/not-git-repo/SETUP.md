# Scenario

**Feature**: wrk --dep errors when dep path is not a git repository

```
# plain directory without .git -> wrk --dep -> non-zero
```

## Steps

1. Create consumer with dep require.
2. Create non-git directory as dep path.
3. Run `wrk --dep`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initConsumerRepo(t, req.WorkRoot, true)
	depPath := filepath.Join(req.WorkRoot, "not-git")
	mkdirAll(t, depPath)
	writeFile(t, filepath.Join(depPath, "go.mod"), "module "+depModulePath+"\n\ngo 1.22\n")

	req.RepoDir = consumer
	req.DepPath = depPath
	req.Args = []string{"--dep", depPath}
	return nil
}
```