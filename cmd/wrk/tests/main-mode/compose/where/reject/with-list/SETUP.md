# Scenario

**Feature**: wrk --main --where --list is mutually exclusive

```
main root -> wrk --main --where --list -> non-zero; mutually exclusive
```

## Steps

1. Parent/leaf creates main repo; cwd = main root.
2. Args = `--main`, `--where`, `--list`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := initMainRepo(t, req)
	req.RepoDir = mainRepo
	req.Args = []string{"--main", "--where", "--list"}
	req.TargetDir = ""
	return nil
}
```
