## Expected

- Non-zero exit code.
- Stderr indicates nothing to commit / no staged changes (intent-stable).
- HEAD subject unchanged.

## Side Effects

- No new commit.

## Errors

- Clean index rejects manual commit.

## Exit Code

- Non-zero

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)

	errText := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	ok := strings.Contains(errText, "nothing to commit") ||
		strings.Contains(errText, "no staged") ||
		strings.Contains(errText, "no changes") ||
		strings.Contains(errText, "nothing staged") ||
		(strings.Contains(errText, "staged") && (strings.Contains(errText, "no ") || strings.Contains(errText, "empty")))
	if !ok {
		t.Fatalf("stderr/stdout should indicate nothing to commit / no staged, got stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}

	subject := gitHEADSubject(t, req.RepoDir)
	if subject != req.HEADSubject {
		t.Fatalf("HEAD subject changed on no-staged fail: before=%q after=%q", req.HEADSubject, subject)
	}
}
```
