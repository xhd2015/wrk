# Scenario

**Feature**: wrk --main --where rejects extra positional (arity 0 with --main)

```
main root -> wrk --main --where foo -> non-zero; unexpected arguments
```

## Steps

1. Initialize main repo; cwd = main root.
2. Args = `--main`, `--where`, `foo`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := initMainRepo(t, req)
	req.RepoDir = mainRepo
	req.Args = []string{"--main", "--where", "foo"}
	req.TargetDir = ""
	return nil
}
```
