# Scenario

**Feature**: embedded wrk-pipeline-publish-and-bump SKILL.md (self-contained)

```
workspace/ -> wrk skill --show wrk-pipeline-publish-and-bump -> stdout that skill's SKILL.md
```

## Steps

1. Run `wrk skill --show wrk-pipeline-publish-and-bump` from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "--show", "wrk-pipeline-publish-and-bump"}
	return nil
}
```
