
## Expected

- Exit code 0.
- Stdout is valid JSON (no ANSI escapes).
- JSON includes planned next tag `v0.0.2` and summary `planned: 1` (or equivalent).
- Tag `v0.0.2` does NOT exist locally.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
	"encoding/json"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	assertNoANSI(t, resp.Stdout)
	obj := assertValidJSONObject(t, resp.Stdout)

	// Accept flexible top-level shapes while requiring the planned bump.
	raw, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("re-marshal json: %v", err)
	}
	body := string(raw)
	assertContains(t, body, "v0.0.2")

	planned := extractJSONPlannedCount(t, obj)
	if planned != 1 {
		t.Fatalf("planned count: want 1, got %d in %v", planned, obj)
	}

	if tagRefExists(t, req.MainRepo, "v0.0.2") {
		t.Fatal("v0.0.2 tag should not exist after JSON dry-run")
	}
}

func extractJSONPlannedCount(t *testing.T, obj map[string]any) int {
	t.Helper()
	if summary, ok := obj["summary"].(map[string]any); ok {
		if n, ok := summary["planned"].(float64); ok {
			return int(n)
		}
	}
	if n, ok := obj["planned"].(float64); ok {
		return int(n)
	}
	if decisions, ok := obj["decisions"].([]any); ok {
		count := 0
		for _, d := range decisions {
			m, ok := d.(map[string]any)
			if !ok {
				continue
			}
			if next, ok := m["next_tag"].(string); ok && next != "" {
				count++
				continue
			}
			if next, ok := m["nextTag"].(string); ok && next != "" {
				count++
			}
		}
		return count
	}
	t.Fatalf("JSON missing planned summary: %v", obj)
	return 0
}
```