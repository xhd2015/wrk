## Expected

- Non-zero exit code.
- Stderr indicates the source has no usable release tags (or cannot resolve source releases).
- Stdout has no successful would-update footer plan (prefer empty stdout).

## Errors

- Hard error when `ResolveSourceReleases` yields no releases for the source main
  (all modules missing numeric tags).

## Side Effects

- Source go.mod / HEAD / tags unchanged (still no tags).

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

	// Must be a real no-release-tags hard error, not flag parse failure.
	// Avoid matching the substring "tag" inside "--propagate-tags".
	lower := strings.ToLower(resp.Stderr)
	if strings.Contains(lower, "unrecognized flag") {
		t.Fatalf("expected implemented --propagate-tags no-tags error, got flag parse failure: %q", resp.Stderr)
	}
	hasHint := strings.Contains(lower, "no release") ||
		strings.Contains(lower, "no numeric") ||
		strings.Contains(lower, "no source release") ||
		strings.Contains(lower, "without a release") ||
		strings.Contains(lower, "missing release") ||
		strings.Contains(lower, "no usable") ||
		(strings.Contains(lower, "release") && (strings.Contains(lower, "missing") || strings.Contains(lower, "none") || strings.Contains(lower, "no "))) ||
		(strings.Contains(lower, "tag") && !strings.Contains(lower, "propagate-tags") &&
			(strings.Contains(lower, "no ") || strings.Contains(lower, "missing") || strings.Contains(lower, "none")))
	if !hasHint {
		t.Fatalf("stderr should mention missing release tags (not merely --propagate-tags), got %q", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "would: update 1") || strings.Contains(resp.Stdout, "would: update 2") {
		t.Fatalf("should not plan positive updates without source tags, stdout=%q", resp.Stdout)
	}
	assertDryRunNoMutation(t, req)
}
```

