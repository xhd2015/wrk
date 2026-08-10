## Expected

- Exit code 0.
- Free-first cascade side effects on the single main:
  1. Path-scope tag `pkgs/shared/v0.0.2` exists (shared leaf tagged first).
  2. Root `go.mod` requires `example.com/root/shared` at `v0.0.2`.
  3. **Keep local replace** for shared (`replace … => ./pkgs/shared` still present).
  4. Cascade pin commit on history: subject prefix `wrk: cascade pin ` mentioning
     the shared module / version.
  5. `go.mod`/`go.sum` committed clean after pin (selective cascade commit).
  6. **Commit-before-tag:** root tag `v0.0.2` tip has the pin commit as ancestor.
  7. Bare origin has main + tags published (`--push` when no pending modules).

## Side Effects

- Module cascade mutates main: tags, pin commit, require bump; does **not** drop
  local replace.
- No TagNextAll-only path that tags root before pin commit lands.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("apply cascade single-repo: want exit 0; exit=%d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if req.MainRepo == "" {
		t.Fatal("MainRepo required")
	}

	// 1) Free-first: shared path-scope tag exists.
	if !tagRefExists(t, req.MainRepo, cascadeSharedNextTag) {
		t.Fatalf("missing free-first shared tag %s on main\nstdout:\n%s\nstderr:\n%s",
			cascadeSharedNextTag, resp.Stdout, resp.Stderr)
	}

	// 2–3) Require bumped + keep local replace (C-AP1 / C-AP3).
	assertRequireBumpedKeepReplace(t, req)

	// 4–5) Cascade pin commit + go.mod clean (not left as WIP pin).
	assertCascadePinCommitPresent(t, req.MainRepo, cascadeSharedModule, unwindApplyNextTag)
	assertGoModCommittedClean(t, req.MainRepo)

	// 6) Commit-before-tag: root next tag only after pin commit (C-AP4).
	if !tagRefExists(t, req.MainRepo, unwindApplyNextTag) {
		t.Fatalf("missing root tag %s after cascade (consumer tag after pin)\nlog:\n%s",
			unwindApplyNextTag, gitOutputIsolated(t, req.MainRepo, "log", "--oneline", "-20"))
	}
	assertCommitBeforeTag(t, req.MainRepo, unwindApplyNextTag)

	// 7) Push: origin main matches local; tags on bare.
	if req.OriginBare == "" {
		t.Fatal("OriginBare required")
	}
	assertOriginMainEqualsLocalMain(t, req.MainRepo, req.OriginBare)
	if !remoteTagExists(t, req.OriginBare, cascadeSharedNextTag) {
		t.Fatalf("origin missing shared tag %s after --push", cascadeSharedNextTag)
	}
	if !remoteTagExists(t, req.OriginBare, unwindApplyNextTag) {
		t.Fatalf("origin missing root tag %s after --push", unwindApplyNextTag)
	}
}
```

