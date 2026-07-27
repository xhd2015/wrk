# Scenario

**Feature**: wrk --main rejects extra positional arguments

```
wrk --main some-path -> non-zero; unexpected arguments
```

## Steps

1. Initialize main repo; cwd = main root (git ok so error is about args).
2. Run `wrk --main some-path`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := initMainRepo(t, req)
	req.RepoDir = mainRepo
	req.Args = []string{"--main", "some-path"}
	req.TargetDir = ""
	return nil
}
```
