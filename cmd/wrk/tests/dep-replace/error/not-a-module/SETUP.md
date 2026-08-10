# Scenario

**Feature**: --dep-replace fails when dep dir is not a go module

```
consumer present; plain dir without go.mod
  -> wrk --dep-replace <plain>
  -> non-zero; go.mod unchanged
```

## Steps

1. Seed consumer.
2. Create a non-module directory and pass it as the dep arg.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupConsumerWithDep(t, req, true)
	plain := filepath.Join(req.WorkRoot, "plain-not-mod")
	mkdirAll(t, plain)
	writeFile(t, filepath.Join(plain, "readme.txt"), "not a module\n")
	plain = resolvePath(t, plain)
	req.DepDir = plain
	req.Args = []string{"--dep-replace", plain}
	return nil
}
```
