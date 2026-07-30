# Scenario

**Feature**: wrk --main --where --exec is invalid (where rejects exec)

```
main root -> wrk --main --where --exec true -> non-zero; --exec not valid with --where
```

## Steps

1. Initialize main repo; cwd = main root.
2. Args = `--main`, `--where`, `--exec`, `true`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := initMainRepo(t, req)
	req.RepoDir = mainRepo
	req.Args = []string{"--main", "--where", "--exec", "true"}
	req.TargetDir = ""
	return nil
}
```
