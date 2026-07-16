## Expected Output

```
source: /abs/.../lib
  example.com/lib  @ v1.2.3  (tag v1.2.3)

would: update 0 modules across 0 projects
```

## Expected

- Exit code 0.
- Source block lists lib release.
- No `would: update` module blocks for `tool` or any consumer.
- Footer zeros.
- Stderr empty.

## Side Effects

- Source go.mod / tags / HEAD unchanged.

## Exit Code

- 0

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}

	var b strings.Builder
	b.WriteString(sourceHeader(req.SourcePath))
	b.WriteByte('\n')
	b.WriteString(sourceReleaseLine("example.com/lib", "v1.2.3", "v1.2.3"))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(planFooter(0, 0))
	b.WriteByte('\n')
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(b.String()))

	if strings.Contains(resp.Stdout, "would: update example.com/") {
		// Footer line also starts with "would: update 0 modules" — allow that.
		// Fail only if a module-path would-update header appears.
		for _, line := range strings.Split(resp.Stdout, "\n") {
			if strings.HasPrefix(line, "would: update example.com/") {
				t.Fatalf("unexpected would-update module block: %q", line)
			}
		}
	}
	assertDryRunNoMutation(t, req)
}
```
