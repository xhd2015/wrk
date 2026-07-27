# Scenario

**Feature**: wrk --all-deps rejects the removed --scan-root flag

```
# wrk --all-deps --scan-root X -> non-zero; stderr mentions unknown/unexpected flag or scan-root removed
consumer + --scan-root <path> -> error
```

## Steps

1. Create a consumer git repo.
2. Run `wrk --all-deps --scan-root <tmpdir>` from the consumer.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	allDepsEnsureHelpersUsed()

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1"}, "")
	scanRoot := filepath.Join(req.WorkRoot, "scan-root")
	mkdirAll(t, scanRoot)

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps", "--scan-root", scanRoot}
	return nil
}
```