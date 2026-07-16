## Expected

- Exit code 0.
- Stdout mock B uses N=2 (binary + text, count before unstage).
- Stderr mentions a planned unstage of the binary (`would` + `unstage` + binary path).
- Binary remains staged after the run (no index mutation).

## Side Effects

- `git diff --cached --name-only` still lists the binary and the text file.
- Agent is not required (dry-run pure plan).

## Exit Code

- 0

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertMockMessageB(t, resp.Stdout, 2)

	binRel := req.BinaryRel
	if binRel == "" {
		binRel = "blob.bin"
	}
	stderrLower := strings.ToLower(resp.Stderr)
	if !strings.Contains(stderrLower, "would") || !strings.Contains(stderrLower, "unstage") {
		t.Fatalf("stderr should plan unstage with would/unstage, stderr:\n%s", resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, binRel) {
		t.Fatalf("stderr should mention binary %q, stderr:\n%s", binRel, resp.Stderr)
	}

	staged := gitStagedNames(t, req.RepoDir)
	joined := strings.Join(staged, "\n")
	if !strings.Contains(joined, binRel) {
		t.Fatalf("binary %q must remain staged after dry-run, staged=%v", binRel, staged)
	}
	if !strings.Contains(joined, "app.go") {
		t.Fatalf("text file app.go must remain staged, staged=%v", staged)
	}
}
```
