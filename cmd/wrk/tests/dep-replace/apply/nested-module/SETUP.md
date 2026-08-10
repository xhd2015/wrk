# Scenario

**Feature**: workDir nested under consumer still edits nearest parent go.mod (D6)

```
consumer/go.mod; workDir = consumer/sub (no go.mod)
  -> wrk --dep-replace <dep>
  -> replace written to consumer/go.mod (not a new go.mod under sub)
  -> exit 0
```

## Steps

1. Seed nested workDir fixture.
2. Run apply with RepoDir = sub.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupNestedConsumerWorkDir(t, req)
	req.Args = []string{"--dep-replace", req.DepDir}
	return nil
}
```
