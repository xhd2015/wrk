# Scenario

**Feature**: sibling out-of-tree dep via `replace => ../external/dep-…`; both dirty

```
# layout:
#   {WorkRoot}/root/                              consumer (RepoDir, main)
#   {WorkRoot}/external/dot-pkgs-main-2026-06-30/ dep (sibling, not nested under root)
# root go.mod: require + replace => ../external/dot-pkgs-main-2026-06-30
# both dirty; pin flags present (synthetic edge once follow lands)
sibling replace both dirty
  -> wrk --unwind --dry-run --tag-next --push
  -> would: peel ../external/dot-pkgs-main-2026-06-30
  -> would: peel .
  -> exit 0; zero mutations
```

## Steps

1. Seed consumer main + sibling dep under `{WorkRoot}/external/…`.
2. Local filesystem replace from consumer to dep; dirtify both.
3. Dry-run with pin flags; PeelOrder = dep display then `.`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupFollowSiblingBothDirty(t, req)
	req.Args = []string{"--unwind", "--dry-run", "--tag-next", "--push"}
	recordUnwindBaseline(t, req)
	return nil
}
```
