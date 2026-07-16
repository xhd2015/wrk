## Expected

- `HTTPStatus` is 200.
- Body is JSON with key `projects` equal to an **empty array** (never null).
- Stdout still printed the listen URL (trailing `\n`).

## Side Effects

- Server process terminated by harness after probe.

## Exit Code

- Ignored after kill.

```go
import (
	"encoding/json"
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.HTTPStatus != 200 {
		t.Fatalf("GET /api/wrk/projects expected 200, got %d body=%q stdout=%q stderr=%q",
			resp.HTTPStatus, resp.HTTPBody, resp.Stdout, resp.Stderr)
	}
	// Decode with json.RawMessage so we can detect null vs [].
	var envelope struct {
		Projects json.RawMessage `json:"projects"`
	}
	if err := json.Unmarshal([]byte(resp.HTTPBody), &envelope); err != nil {
		t.Fatalf("invalid JSON body: %v body=%q", err, resp.HTTPBody)
	}
	if len(envelope.Projects) == 0 {
		t.Fatalf("missing projects key; body=%q", resp.HTTPBody)
	}
	if string(envelope.Projects) == "null" {
		t.Fatalf("projects must be [] not null; body=%q", resp.HTTPBody)
	}
	var list []any
	if err := json.Unmarshal(envelope.Projects, &list); err != nil {
		t.Fatalf("projects is not an array: %v body=%q", err, resp.HTTPBody)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty projects array, got %d entries; body=%q", len(list), resp.HTTPBody)
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("stdout must end with trailing newline; got %q", resp.Stdout)
	}
	assert.Output(t, resp.Stdout, `---
version: 3
__PORT__: type=number, example=18080, TCP listen port
---
http://127\.0\.0\.1:__PORT__/
`)
}
```
