# Scenario

**Feature**: wrk --gen-commit-msg -h documents mode and key flags

```
workspace/ -> wrk --gen-commit-msg -h
  -> exit 0
  -> help mentions gen-commit-msg options:
     --gen-commit-msg / gen-commit-msg, --model, --dry-run, --commit, --no-verify, --agent-runner
```

## Steps

1. Run `wrk --gen-commit-msg -h` from neutral cwd (no git required).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--gen-commit-msg", "-h"}
	return nil
}
```
