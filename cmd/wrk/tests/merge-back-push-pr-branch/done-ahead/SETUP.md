# Scenario

**Feature**: `--done --push` lease-updates existing `origin/<WtBranch>` after land+remove

```
# --done deletes the local branch; still publish post-land SHA to origin/<WtBranch>
  -> wrk --done -y --push
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupAheadWithOriginBranch(t, req)
	req.Args = []string{"--done", "-y", "--push"}
	return nil
}
```
