## Expected Output

```
==== unwind graph (repo) ====
  … dirty …
  peel order (dirty, free-first):
    .

==== unwind graph (module) ====
  dir  …
  .

==== status summary ====
  …
```

## Expected

- Exit code 0.
- Human banners present.
- Peel order includes display `.` (not bare basename `root` alone as peel path).
- Module identity dir `.` listed (not required as full `example.com/root` path).
- Repo node dirty status language present (`dirty`).
- No flat full-path module edges.
- Zero mutations: HEAD unchanged; DIRTY still present.

## Side Effects

- None.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertShowGraphHumanBanners(t, resp.Stdout)
	assertShowGraphPeelOrderHuman(t, resp.Stdout, req.PeelOrder)
	assertModuleDirListed(t, resp.Stdout, ".")
	assertHumanNoFullModulePaths(t, resp.Stdout)
	assertHumanNoFlatFullPathEdges(t, resp.Stdout)
	if !strings.Contains(strings.ToLower(resp.Stdout), "dirty") {
		t.Fatalf("dirty single-main graph should mention dirty; stdout:\n%s", resp.Stdout)
	}
	assertShowGraphZeroMutations(t, req)
	assertFileExists(t, filepath.Join(req.MainRepo, "DIRTY"))
}
```
