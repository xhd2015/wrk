## Expected Output

```json
{
  "repos": {
    "peel_order": [
      "external/dot-pkgs-main-2026-06-30",
      "external/agent-pro-main-2026-06-30",
      "."
    ],
    …
  },
  …
}
```

## Expected

- Exit code 0.
- JSON shape keys present.
- `repos.peel_order` exactly matches free-first display path sequence.
- Zero mutations.

## Side Effects

- None.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	_, repos, _ := assertGraphJSONShape(t, resp.Stdout)
	assertJSONPeelOrder(t, repos.PeelOrder, req.PeelOrder)
	// Multi dirty with linked + edges: needs_land / has_pending_edges should be true.
	if repos.NeedsLand != nil && !*repos.NeedsLand {
		t.Fatalf("linked dirty stack: needs_land should be true; peel=%#v", repos.PeelOrder)
	}
	if repos.HasPendingEdges != nil && !*repos.HasPendingEdges {
		t.Fatalf("3-repo chain with edges: has_pending_edges should be true")
	}
	assertShowGraphZeroMutations(t, req)
}
```
