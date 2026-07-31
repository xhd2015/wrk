# Scenario

**Feature**: multi-bring resolves two registered basenames via projects.json

```
# saved mydep1 + mydep2 in projects.json; no local cwd copies
# wrk --bring mydep1 --bring mydep2
#   -> both resolve + external worktrees + replaces; exit 0
consumer requires both modules
  -> multi-bring basenames
```

## Steps

1. Create consumer requiring both modules (with imports).
2. Create saved dep repos and `wrk --add` each.
3. Run `wrk --bring mydep1 --bring mydep2` from consumer.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initMultiBringConsumerWithTwoRequires(t, req.WorkRoot)

	// Nested under saved/ so filepath.Base is mydep1/mydep2 for projects.json match.
	dep1 := initMultiBringDepRepo(t, req.WorkRoot, filepath.Join("saved", "mydep1"), multiBringDep1Module)
	dep2 := initMultiBringDepRepo(t, req.WorkRoot, filepath.Join("saved", "mydep2"), multiBringDep2Module)
	multiRecordSavedProject(t, req, dep1)
	multiRecordSavedProject(t, req, dep2)

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.DepPath = dep1
	req.SecondRepo = dep2
	req.Args = []string{"--bring", "mydep1", "--bring", "mydep2"}
	return nil
}
```
