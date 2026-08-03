# Scenario

**Feature**: bare --commit without gen or -m needs a message source

```
workspace/ -> wrk --commit
  -> non-zero; needs -m/--message or --gen-commit-msg
```

## Preconditions

- Locked decision D3: bare `--commit` without gen/`-m` → hard error.

## Steps

1. Run `wrk --commit` from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--commit"}
	return nil
}
```
