## Expected Output

```
==== dep-replace ====
dep  example.com/dep => <abs>

  checkout  ..
    module  example.com/consumer
      replace  example.com/dep => <abs>
      go mod tidy ok

dep-replace: replaced in 1 modules in 1 checkouts
```

## Expected

- Exit 0.
- Checkout label is `..` (`statusDirLine` of consumer vs nested cwd).
- Replace lands in **consumer** go.mod (parent), not under `sub/`.
- No go.mod created under nested workDir.

## Side Effects

- Nearest go.mod above workDir receives absolute replace (not-git D6).

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assert.Output(t, resp.Stdout, `---
version: 3
__ABS__: type=string
---
==== dep-replace ====
dep  example\.com/dep => __ABS__

  checkout  \.\.
    module  example\.com/consumer
      replace  example\.com/dep => __ABS__
      go mod tidy ok

dep-replace: replaced in 1 modules in 1 checkouts
`)
	assertAbsoluteReplace(t, req.ConsumerGoMod, modDep, req.DepDir)
	nestedGoMod := filepath.Join(req.RepoDir, "go.mod")
	if _, err := os.Stat(nestedGoMod); err == nil {
		t.Fatalf("must not create go.mod under nested workDir %s", req.RepoDir)
	}
}
```
