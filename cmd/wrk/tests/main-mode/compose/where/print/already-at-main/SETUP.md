# Scenario

**Feature**: wrk --main --where at main root still prints path (no bare-main notice)

```
mainRepo (cwd) -> wrk --main --where
  -> stdout mainRepo\n
  -> empty stderr (no "already at main repository root")
  -> no shell
```

## Steps

1. Initialize main repo; cwd = main root.
2. Install fake bash.
3. Run `wrk --main --where`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := initMainRepo(t, req)
	req.RepoDir = mainRepo
	installFakeBash(t, req, 0)
	setMainWhereArgs(req, "--main", "--where")
	return nil
}
```
