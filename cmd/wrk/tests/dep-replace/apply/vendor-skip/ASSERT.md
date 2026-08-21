## Expected Output

```
==== dep-replace ====
dep  example.com/dep => <abs>

  checkout  .
    module  example.com/consumer
      replace  example.com/dep => <abs>
      skip tidy  (vendor/)

dep-replace: replaced in 1 modules in 1 checkouts
```

## Expected

- Exit 0.
- Absolute replace applied.
- `skip tidy  (vendor/)` under the module; no `go mod tidy ok`.
- No go.sum; vendor/ untouched.

## Side Effects

- Replace only; tidy skipped because of vendor/.

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

  checkout  \.
    module  example\.com/consumer
      replace  example\.com/dep => __ABS__
      skip tidy  \(vendor/\)

dep-replace: replaced in 1 modules in 1 checkouts
`)
	assertNotContains(t, resp.Stdout, "go mod tidy ok")
	assertAbsoluteReplace(t, req.ConsumerGoMod, modDep, req.DepDir)
	assertNoTidyArtifacts(t, req)
	if req.VendorDir == "" {
		t.Fatal("fixture bug: VendorDir unset")
	}
	if _, err := os.Stat(filepath.Join(req.VendorDir, "modules.txt")); err == nil {
		t.Fatalf("vendor/ must not gain modules.txt")
	}
}
```
