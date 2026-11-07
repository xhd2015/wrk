# Scenario

**Feature**: wrk skill --install <name> --cursor --dry-run installs that named skill

```
embedded skills in wrk binary
workspace/ -> wrk skill --install wrk-pipeline-publish-and-bump --cursor --dry-run
  -> dry-run lines for .cursor/skills/wrk-pipeline-publish-and-bump
```

## Steps

1. Run `wrk skill --install wrk-pipeline-publish-and-bump --cursor --dry-run`
   from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "--install", "wrk-pipeline-publish-and-bump", "--cursor", "--dry-run"}
	return nil
}
```
