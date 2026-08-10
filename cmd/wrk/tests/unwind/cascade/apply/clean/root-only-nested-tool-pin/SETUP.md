# Scenario

**Bug / gap**: root-only tagscope + pin-only nested tool module (ai-critic
`script/browser-debug` shape) — cascade dies with `tag already exists`

```
# single main: only bare tags v0.0.1 (no path-scoped nested release tags)
# root example.com/root owned-changed → next v0.0.2
# nested script/browser-debug (module browser-debug):
#   require root@v0.0.1 + replace => ../../  (keep-local)
# no pkgs/shared path tags — tool is pin consumer only
stack already on main (DIRTY peel marker)
  -> wrk --unwind --tag-next --push
  -> desired: tag-next root @ v0.0.2
  ->         pin browser-debug <- root @ v0.0.2; KEEP replace
  ->         NO tag-next for browser-debug
  ->         exit 0; push main + tag
  -> classic bug: planner attaches root NextTag to browser-debug →
       tag root @ v0.0.2 → pin commit moves HEAD →
       git tag v0.0.2 again → fatal: tag 'v0.0.2' already exists
```

## Steps

1. Seed `setupApplyCascadeRootOnlyNestedToolPin` (bare origin, clean go.mod Base).
2. Run non-dry-run `--unwind --tag-next --push`.
3. Assert root tag + tool pin keep-replace; exit 0; no second same-name tag.

## Context

- Reproduces production failure on `ai-critic` after cascade pin of
  `script/browser-debug/go.mod` @ `v0.0.18` while re-tagging `v0.0.18`.
- Distinct from C-AP1 (`pkgs/shared/v…` path scopes — tag names never collide).
- Was Classic **RED** until `attachTagScopeToModules` stopped giving root `NextTag`
  to nested modules that have no matching path-scope decision.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyCascadeRootOnlyNestedToolPin(t, req)
	req.Args = []string{"--unwind", "--tag-next", "--push"}
	return nil
}
```
