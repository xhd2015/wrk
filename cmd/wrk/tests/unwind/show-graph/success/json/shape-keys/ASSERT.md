## Expected Output

Pure JSON (no ANSI, no human banners):

```json
{
  "repos": {
    "nodes": […],
    "edges": […],
    "peel_order": ["."],
    "has_pending_edges": false,
    "needs_land": false
  },
  "modules": {
    "nodes": […],
    "edges": […]
  },
  "summary": {…},
  "warnings": […]
}
```

(Exact nested field sets implementer-owned beyond locked keys.)

## Expected

- Exit code 0.
- Stdout parses as JSON with required keys.
- `repos.peel_order` is an array (includes `.` for dirty single main).
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
	assertShowGraphZeroMutations(t, req)
}
```
