## Expected

- Root probe performed HTTP GET `/mockup/repo-view`.
- `HTTPStatus` is 200.
- Body is the SPA HTML shell (DOCTYPE/html, root mount, wrk markers).
- Client route is not a bare 404 (same shell as `/` for SPA fallback).

## Side Effects

- Server process is terminated by the harness after the probe.

## Exit Code

- Ignored after kill.

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.HTTPStatus != 200 {
		t.Fatalf("GET /mockup/repo-view expected 200, got %d body=%q stdout=%q stderr=%q",
			resp.HTTPStatus, resp.HTTPBody, resp.Stdout, resp.Stderr)
	}
	body := resp.HTTPBody
	lower := strings.ToLower(body)
	if !strings.Contains(lower, "<html") && !strings.Contains(lower, "<!doctype") {
		t.Fatalf("expected HTML body, got %q", body)
	}
	if !strings.Contains(body, "root") && !strings.Contains(body, "id=\"root\"") {
		// SPA must mount on #root
		if !strings.Contains(lower, "id=\"root\"") && !strings.Contains(body, "id='root'") {
			t.Fatalf("expected SPA root mount in body, got %q", body)
		}
	}
	if !strings.Contains(body, "wrk") {
		t.Fatalf("expected wrk marker in SPA shell, got %q", body)
	}
	// Static fallback / shell still carries diagram keywords for non-JS probes.
	for _, marker := range []string{"task", "Main", "Remote"} {
		if !strings.Contains(body, marker) {
			t.Fatalf("SPA shell missing marker %q; body=%q", marker, body)
		}
	}
}
```
