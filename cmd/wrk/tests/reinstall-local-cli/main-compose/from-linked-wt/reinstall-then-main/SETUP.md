# Scenario

**Feature**: --reinstall-local --main --dry-run same plan as main-first (MC3)

```
# MC3: flag order free — reinstall-local before --main
linked-wt -> wrk --reinstall-local --main --dry-run
  -> same main multi plan as --main --reinstall-local --dry-run
```

## Steps

1. Parent built diverged main + linked-wt; cwd = linked-wt.
2. Override Args to `--reinstall-local --main --dry-run`.
3. Expect identical stdout to MC1.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--reinstall-local", "--main", "--dry-run"}
	return nil
}
```
