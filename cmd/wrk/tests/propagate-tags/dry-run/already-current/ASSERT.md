## Expected Output

```
source: /abs/.../lib
  example.com/lib  @ v1.2.3  (tag v1.2.3)

would: update 0 modules across 0 projects
```

## Expected

- Exit code 0.
- Source release block is printed.
- Stdout has **no** `would: update example.com/app` module block.
- Footer is zero modules / zero projects (or an explicit already-current count
  may appear in addition only if it does not invent a would-update module block —
  this leaf locks the zero-footer form).

## Side Effects

- go.mod / tags / HEAD unchanged on lib and app.

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

	if strings.Contains(resp.Stdout, "would: update example.com/app") {
		t.Fatalf("already-current consumer must not have would-update block, got %q", resp.Stdout)
	}
	assertDryRunNoMutation(t, req)
}
```
