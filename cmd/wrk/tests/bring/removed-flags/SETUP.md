# Scenario

**Feature**: legacy `--dep` and `--all-deps` are hard-removed from the CLI surface

```
# end-state: only --bring remains as external-dep worktree mode
wrk --dep <path> -> non-zero; unknown/invalid flag --dep
wrk --all-deps -> non-zero; unknown/invalid flag --all-deps
# help / dry-run host lists no longer advertise the removed flags
```

## Steps

- Leaves use minimal cwd fixtures and L2 `InProcess` for fast-fail unknown-flag paths.
- Asserts prefer stable substrings: flag name + `unknown` / invalid-style messaging.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	_ = helpLineDocumentsFlag
	return nil
}

// helpLineDocumentsFlag reports whether a help line documents flag as its own
// token (not a longer flag that shares a prefix, e.g. --dep vs --dep-replace).
func helpLineDocumentsFlag(trimLine, flag string) bool {
	if trimLine == flag {
		return true
	}
	if strings.HasPrefix(trimLine, flag+" ") || strings.HasPrefix(trimLine, flag+"\t") || strings.HasPrefix(trimLine, flag+"=") {
		return true
	}
	// common usage indent: "  --dep " / "  --dep\t"
	if strings.Contains(trimLine, " "+flag+" ") || strings.Contains(trimLine, "\t"+flag+" ") ||
		strings.Contains(trimLine, " "+flag+"\t") || strings.HasSuffix(trimLine, " "+flag) {
		return true
	}
	return false
}
```
