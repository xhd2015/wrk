## Expected

- `err` is nil.
- `resp.Tags` has length ≥ 2 (root + nested scope planned names).
- Planned names include `v0.0.2` and `sub/v0.2.4` for this seed.
- `resp.Tag` (primary) is non-empty.
- Tag refs `v0.0.2` and `sub/v0.2.4` do **not** exist (DryRun only).
- `resp.TagMainRepo` equals main repo when MainRepo is populated.

## Side Effects

- None (plan only; no lightweight tags created for any scope).

## Errors

- None.

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	if len(resp.Tags) < 2 {
		t.Fatalf("Tags length: got %d want ≥ 2 (multi-scope plan); Tags=%v", len(resp.Tags), resp.Tags)
	}
	if strings.TrimSpace(resp.Tag) == "" {
		t.Fatal("Tag (primary) empty after TagNextAll DryRun")
	}
	wantRoot := "v0.0.2"
	wantSub := "sub/v0.2.4"
	if !containsString(resp.Tags, wantRoot) {
		t.Fatalf("Tags missing %q: %v", wantRoot, resp.Tags)
	}
	if !containsString(resp.Tags, wantSub) {
		t.Fatalf("Tags missing %q: %v", wantSub, resp.Tags)
	}
	if tagRefExists(t, req.MainRepo, wantRoot) {
		t.Fatalf("%s tag should not exist after DryRun", wantRoot)
	}
	if tagRefExists(t, req.MainRepo, wantSub) {
		t.Fatalf("%s tag should not exist after DryRun", wantSub)
	}
	// MainRepo from TagNextResult should resolve to the fixture main when set.
	if resp.TagMainRepo != "" && resolvePath(t, resp.TagMainRepo) != resolvePath(t, req.MainRepo) {
		t.Fatalf("TagMainRepo: got %q want %q", resp.TagMainRepo, req.MainRepo)
	}
}

func containsString(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}
```
