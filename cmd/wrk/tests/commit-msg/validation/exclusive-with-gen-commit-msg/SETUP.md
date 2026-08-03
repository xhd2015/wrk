# Scenario

**Feature**: --gen-commit-msg and -m are mutually exclusive

```
workspace/ -> wrk --gen-commit-msg --commit --message "feat: clash"
  -> non-zero; mutually exclusive / exclusive
```

## Preconditions

- Message sources are XOR: AI gen vs manual `-m`/`--message`. Combining them must fail early.
- Use long form `--message` so the pin cannot be confused with a short alias for `--model`.

## Steps

1. Run `wrk --gen-commit-msg --commit --message "feat: clash"` from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--gen-commit-msg", "--commit", "--message", "feat: clash"}
	return nil
}
```

