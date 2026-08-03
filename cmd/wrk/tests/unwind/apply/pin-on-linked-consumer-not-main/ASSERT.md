## Expected

- Exit code 0.
- Leaf main advanced with feature content; local tag `v0.0.2` at leaf main HEAD;
  bare origin has main + tag (dep ship still uses dep main — unchanged by this rule).
- **Linked consumer Path** (`req.WtDir`) `go.mod`:
  - requires `example.com/dot-pkgs` at `v0.0.2`
  - **no** local replace for that module
- **Consumer MainRepo** (`req.MainRepo`) `go.mod`:
  - **byte-equal** to pre-run baseline snapshot
  - still requires `example.com/dot-pkgs` at `v0.0.1` (never pinned)
- Path ≠ MainRepo (fixture invariant).

## Side Effects

- Peel lands nested leaf; tag-next + push on leaf main.
- Pin must edit only the in-scope consumer Path, not the out-of-scope MainRepo checkout.
- Log may print `pin root <- dot-pkgs @ v0.0.2` (basename); do not treat log path as surface.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		combined := resp.Stdout + "\n" + resp.Stderr
		if strings.Contains(combined, "not implemented") {
			t.Fatalf("apply not implemented yet: exit=%d stderr=%q stdout=%q",
				resp.ExitCode, resp.Stderr, resp.Stdout)
		}
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if req.WtDir == "" || req.MainRepo == "" {
		t.Fatal("WtDir (linked Path) and MainRepo required")
	}
	wt := resolvePath(t, req.WtDir)
	main := resolvePath(t, req.MainRepo)
	if wt == main {
		t.Fatal("fixture must have linked consumer Path ≠ MainRepo")
	}
	if resolvePath(t, req.RepoDir) != wt {
		t.Fatalf("RepoDir must be linked consumer Path; RepoDir=%s WtDir=%s", req.RepoDir, req.WtDir)
	}

	// Dep ship still OK on leaf main (not the pin-path rule under test).
	assertLeafMainAdvancedAndTagged(t, req)

	// Key contract: pin Path, not MainRepo.
	assertLinkedConsumerPathPinned(t, req)
	assertConsumerMainGoModUnchanged(t, req)
}
```
