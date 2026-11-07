# Scenario

**Feature**: wrk skill -l is the short alias for --list

```
embedded skills in wrk binary (wrk, wrk-pipeline,
wrk-pipeline-publish-and-bump, edit-dependency-code)
workspace/ -> wrk skill --list -> 4 "name - description" lines
```

## Steps

1. Run `wrk skill -l` from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "-l"}
	return nil
}
```
