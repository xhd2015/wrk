# Scenario

**Feature**: clean leaf omitted from pending; peel mid external then primary `.`

```
# same 3-repo chain; only agent-pro + root dirty; dot-pkgs clean
dot-pkgs (clean) · agent-pro (dirty) · root (dirty)
  -> wrk --unwind --dry-run --tag-next --push --done
  -> would: peel external/agent-pro-main-2026-06-30
  -> would: peel .
  -> no would: peel external/dot-pkgs-…
```

## Steps

1. Build three-repo chain.
2. Dirtify mid + root only; leave leaf clean.
3. Run unwind dry-run with pin + land flags.
4. PeelOrder = mid external display + primary `.`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupThreeRepoChain(t, req)
	dirtyMidAndRoot(t, req)
	setPeelOrderDisplays(t, req, req.ExternalWtDir, req.WtDir)
	req.Args = []string{"--unwind", "--dry-run", "--tag-next", "--push", "--done"}
	recordUnwindBaseline(t, req)
	return nil
}
```
