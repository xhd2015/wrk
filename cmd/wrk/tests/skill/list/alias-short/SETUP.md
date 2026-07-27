# Scenario

**Feature**: wrk skill -l is the short alias for --list

```
embedded SKILL.md in wrk binary
workspace/ -> wrk skill -l -> stdout wrk\n
```

## Steps

1. Run `wrk skill -l` from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "-l"}
	return nil
}
```
