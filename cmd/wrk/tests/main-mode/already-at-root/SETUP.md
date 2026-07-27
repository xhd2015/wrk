# Scenario

**Feature**: wrk --main at main repo root prints notice and does not launch a shell

```
# cwd is main checkout root
mainRepo (cwd) -> wrk --main
  -> stderr: already at main repository root: <mainRepo>
  -> no LoginInteractive
  -> exit 0; empty stdout
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo` on branch `main`.
2. Set process cwd to that main root.
3. Install fake bash so an accidental shell launch is observable (not hung).
4. Run `wrk --main`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := initMainRepo(t, req)
	req.RepoDir = mainRepo
	setMainArgs(req)
	// Detect accidental shell: if implementation wrongly launches, log is non-empty.
	// Under Capture the fake PATH is not applied; non-launch still leaves log empty.
	installFakeBash(t, req, 0)
	return nil
}
```
