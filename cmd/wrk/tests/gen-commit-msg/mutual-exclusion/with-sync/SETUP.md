# Scenario

**Feature**: bare `--gen-commit-msg --sync` is allowed multi-stage compose (activeRoot stays cwd; no done required)

```
# Target model: pipeline stages may compose without --done/--merge-back
repo (staged change)
  -> wrk --gen-commit-msg --sync --dry-run
  -> must NOT stderr "mutually exclusive"
  -> exit 0: gen-commit dry plan + sync dry plan (may be empty summary)
```

## Steps

1. Stage one text file via `stageOneTextFile` (nested gen-commit-msg helpers).
2. Run `wrk --gen-commit-msg --sync --dry-run` from that git repo.

```go
func Setup(t *testing.T, req *Request) error {
	stageOneTextFile(t, req)
	req.Args = []string{"--gen-commit-msg", "--sync", "--dry-run"}
	return nil
}
```
