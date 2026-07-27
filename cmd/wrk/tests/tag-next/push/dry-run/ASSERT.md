
## Expected Output

```
v0.0.1        owned changed                  ->  v0.0.2
1 tag planned

would: git push origin main
would: git push origin v0.0.2
```

## Expected

- Exit code 0.
- Tag-next dry-run plan for root bump to `v0.0.2`.
- Push dry-run lists branch then planned tag (`would: git push origin …`).
- No local tag `v0.0.2`; origin main tip unchanged; no origin tag `v0.0.2`.
- No human `pushed …` confirmation line.

## Side Effects

- None (plan only).

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/gitops/git/git_isolated"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	if tagRefExists(t, req.MainRepo, "v0.0.2") {
		t.Fatal("v0.0.2 must not exist locally under --dry-run")
	}
	// Avoid show-ref --tags MustOutput when origin has zero tags (exit 1).
	if git_isolated.Command(req.OriginBare, "rev-parse", "--verify", "refs/tags/v0.0.2").Run() == nil {
		t.Fatal("v0.0.2 must not exist on origin under --dry-run")
	}

	beforeBytes, err := os.ReadFile(filepath.Join(req.WorkRoot, "origin-main-before"))
	if err != nil {
		t.Fatalf("read origin snapshot: %v", err)
	}
	before := strings.TrimSpace(string(beforeBytes))
	after := strings.TrimSpace(gitOutputIsolated(t, req.OriginBare, "rev-parse", "refs/heads/main"))
	if after != before {
		t.Fatalf("origin/main mutated under --dry-run: before %s after %s", before, after)
	}

	want := "v0.0.1        owned changed                  ->  v0.0.2\n" +
		"1 tag planned\n" +
		"\n" +
		"would: git push origin main\n" +
		"would: git push origin v0.0.2\n"
	assert.Output(t, resp.Stdout, tagNextStdoutV2(want))

	if strings.Contains(resp.Stdout, "pushed ") {
		t.Fatalf("dry-run must not print pushed confirmation; got %q", resp.Stdout)
	}
}
```
