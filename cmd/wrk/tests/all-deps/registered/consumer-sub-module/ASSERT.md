## Expected

- Exit code 0.
- Stdout contains both dep lines and ends with `wrked 2 deps`.
- Consumer sub-module `go-pkgs/go.mod` has replaces for both deps.

## Exit Code

- 0

```go
import (

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"

	"fmt"
	"path/filepath"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantStdout := fmt.Sprintf("wrk example.com/dep1 at %s\nwrk example.com/dep2 at %s\nwrked 2 deps\n",
		allDepsExternalRelPath("mydep1"), allDepsExternalRelPath("mydep2"))
	assert.Output(t, resp.Stdout, allDepsStdoutV2(wantStdout))

	modDir := filepath.Join(req.ConsumerTop, "go-pkgs")
	mod, err := allDepsReadGoMod(modDir)
	if err != nil {
		t.Fatalf("read go-pkgs/go.mod: %v", err)
	}
	wantPath1 := allDepsExternalAbsPath(req.ConsumerTop, "mydep1")
	wantPath2 := allDepsExternalAbsPath(req.ConsumerTop, "mydep2")
	if !allDepsHasReplaceForModule(mod, "example.com/dep1", wantPath1) {
		t.Fatalf("go-pkgs/go.mod missing replace example.com/dep1 => %s: %+v", wantPath1, mod.Replace)
	}
	if !allDepsHasReplaceForModule(mod, "example.com/dep2", wantPath2) {
		t.Fatalf("go-pkgs/go.mod missing replace example.com/dep2 => %s: %+v", wantPath2, mod.Replace)
	}
	ok, err := allDepsGitignoreContainsExternal(req.ConsumerTop)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !ok {
		t.Fatalf(".gitignore should contain /external")
	}
}
```