# Scenario

**Feature**: embedded wrk-pipeline SKILL.md (four phases)

```
workspace/ -> wrk skill --show wrk-pipeline -> stdout that skill's SKILL.md
```

## Steps

1. Run `wrk skill --show wrk-pipeline` from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "--show", "wrk-pipeline"}
	return nil
}
```
