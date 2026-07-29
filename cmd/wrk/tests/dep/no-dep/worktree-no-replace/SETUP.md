# Scenario

**Feature**: --dep --no-dep creates external worktree without replace or tidy

```
# consumer requires dep -> wrk --dep <dep> --no-dep
#   -> external wt + gitignore; go.mod byte-identical (no replace)
consumer (require example.com/dep) + mydep
  -> wrk --dep <dep> --no-dep -> stdout abs path; no replace
```

## Steps

1. Create consumer requiring dep + dep repo.
2. Snapshot go.mod.
3. Run `wrk --dep <dep> --no-dep`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initConsumerRepo(t, req.WorkRoot, true)
	depPath := initDepRepo(t, req.WorkRoot, "mydep", true)

	req.RepoDir = consumer
	req.DepPath = depPath
	req.ConsumerTop = consumer
	req.DepModulePath = depModulePath
	snapshotDepGoMod(t, req, consumer)
	req.Args = []string{"--dep", depPath, "--no-dep"}
	return nil
}
```
