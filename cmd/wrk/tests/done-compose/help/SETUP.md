# Scenario

**Feature**: root help documents full-band composition (gen-commit pre + primary + posts + reinstall tail)

```
# user-facing usage() matches composition: optional gen-commit pre + primary + optional posts/tail
wrk --help
  -> exit 0
  -> --done / --merge-back synopsis lists optional --gen-commit-msg (pre), post modifiers, --reinstall-local
  -> --gen-commit-msg documents pre-stage with --done/--merge-back (and --commit when composed)
  -> --reinstall-local documents validity after primary (and --main)
  -> --push dual meaning (not only with --tag-next)
  -> no claim that --tag-next is exclusive with --done
  -> fluent recipes (or equivalent flag lists) for ship flows:
       wrk --done --sync --tag-next --push -y
       wrk --done --sync --tag-next --push --reinstall-local -y
       wrk --gen-commit-msg --commit --model=… --done --sync --tag-next --push [-y]
```

## Steps

1. Run `wrk --help` from isolated WorkRoot (no git required).

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--help"}
	return nil
}
```
