# Scenario

**Feature**: --main --reinstall-local --dry-run from linked WT plans main modules (MC1)

```
# MC1: cwd=linked-wt (diverged); Args = --main --reinstall-local --dry-run
linked-wt -> wrk --main --reinstall-local --dry-run
  -> useMain=true → scan mainrepo
  -> would: go install ./cmd/mainbin + ./cmd/toolbin (not wtbin)
  -> no nested shell; dry-run only
```

## Steps

1. Parent built diverged main + linked-wt; cwd = linked-wt.
2. Args already defaulted to `--main --reinstall-local --dry-run` by parent.
3. Expect multi dry-run for **main** modules only.

```go
func Setup(t *testing.T, req *Request) error {
	// Parent from-linked-wt sets ModuleRoot=linked-wt and compose Args.
	req.Args = []string{"--main", "--reinstall-local", "--dry-run"}
	return nil
}
```
