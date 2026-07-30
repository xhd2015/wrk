# Scenario

**Feature**: wrk --main --cd /path rejects extra positional (was former mutual-exclusion leaf)

```
main root -> wrk --main --cd /some/path -> non-zero; unexpected arguments
# not "mutually exclusive" — compose is allowed without path; path is the error
```

## Steps

1. Initialize main repo; cwd = main root.
2. Args = `--main`, `--cd`, main path (exists; rejection is arity).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := initMainRepo(t, req)
	req.RepoDir = mainRepo
	req.Args = []string{"--main", "--cd", mainRepo}
	req.TargetDir = ""
	return nil
}
```
