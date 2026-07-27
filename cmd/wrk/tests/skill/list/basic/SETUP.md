# Scenario

**Feature**: wrk skill --list prints wrk on stdout

```
embedded SKILL.md in wrk binary
workspace/ -> wrk skill --list -> stdout wrk\n
```

## Steps

1. Run `wrk skill --list` from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "--list"}
	return nil
}
```
