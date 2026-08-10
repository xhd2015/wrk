## Expected Output

Pure JSON (no ANSI, no human banners):

```json
{
  "work_dir": "...",
  "checks": [
    {"id":"dirty-peel","severity":"error","status":"pass"},
    {"id":"needs-land","severity":"error","status":"pass"},
    {"id":"owned-changed","severity":"error","status":"pass","count":0},
    {"id":"require-drift","severity":"error","status":"pass","count":0},
    {"id":"droppable-replace","severity":"error","status":"pass","count":0},
    {"id":"cascade-pending","severity":"error","status":"pass"}
  ],
  "summary":{"checks":6,"pass":6,"fail":0,"warn":0,"result":"pass"},
  "warnings":[]
}
```

(Exact count fields optional beyond locked keys; all six ids required.)

## Expected

- Exit code 0.
- Stdout parses as JSON with required keys and 6 catalog checks all `pass`.
- `summary.result` is `pass`.
- No ANSI / human banners.
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
	_, checks, sum := assertVerifyJSONShape(t, resp.Stdout)
	for _, id := range verifyCheckIDs {
		assertVerifyJSONCheckStatus(t, checks, id, "pass")
	}
	if sum.Result != "pass" {
		t.Fatalf("summary.result=%q want pass", sum.Result)
	}
	if sum.Fail != 0 {
		t.Fatalf("summary.fail=%d want 0", sum.Fail)
	}
	assertVerifyZeroMutations(t, req)
}
```
