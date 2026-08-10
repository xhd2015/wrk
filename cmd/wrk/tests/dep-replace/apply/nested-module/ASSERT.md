## Expected

- Exit 0.
- Stdout `dep-replace example.com/dep => <abs>`.
- Replace lands in **consumer** go.mod (parent), not under `sub/`.
- No go.mod created under nested workDir.

## Side Effects

- Nearest go.mod above workDir receives absolute replace (D6).

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertDepReplaceLine(t, resp.Stdout, modDep, req.DepDir)
	assertAbsoluteReplace(t, req.ConsumerGoMod, modDep, req.DepDir)
	// Nested workDir must not gain its own go.mod.
	nestedGoMod := filepath.Join(req.RepoDir, "go.mod")
	if _, err := os.Stat(nestedGoMod); err == nil {
		t.Fatalf("must not create go.mod under nested workDir %s", req.RepoDir)
	}
	assertNoTidyArtifacts(t, req)
}
```
