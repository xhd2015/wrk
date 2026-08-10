## Expected Output

```
==== unwind (dry-run) ====
would: peel .
would: tag-next example.com/root/shared @ pkgs/shared/v0.0.2
would: pin example.com/root <- example.com/root/shared @ v0.0.2
```

(No `would: tag-next` for `example.com/root/testdata-x` or any `testdata/` scope.)

## Expected

- Exit code 0.
- Positive cascade: tag-next for **shared** free module present (free-first).
- **No** tag-next line whose module identity or tag mentions testdata scope:
  - module path `example.com/root/testdata-x`
  - tag / path fragment `testdata/x`
- Zero mutations.

## Side Effects

- None.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertPeelOrder(t, resp.Stdout, req.PeelOrder)

	out := resp.Stdout
	// Positive: real free module still cascaded (RED driver today).
	if !hasCascadeTagNext(out, cascadeSharedModule) {
		t.Fatalf("missing tag-next for free shared module %s\nstdout:\n%s",
			cascadeSharedModule, out)
	}

	// Negative: no cascade tag-next for testdata module path or path-scope tags.
	if hasCascadeTagNext(out, cascadeTestdataModule) {
		t.Fatalf("testdata module must not get would: tag-next\nstdout:\n%s", out)
	}
	for _, raw := range strings.Split(out, "\n") {
		if !strings.HasPrefix(strings.TrimLeft(raw, " \t"), "would: tag-next ") {
			continue
		}
		lower := strings.ToLower(raw)
		if strings.Contains(lower, "testdata") {
			t.Fatalf("cascade tag-next must not mention testdata scope: %q\nstdout:\n%s",
				raw, out)
		}
	}
	assertUnwindZeroMutations(t, req)
}
```
