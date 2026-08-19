# Scenario

**Feature**: `--merge-back --push --dry-run` plans lease-update of existing `origin/<WtBranch>` without mutating

```
  -> wrk --merge-back --push --dry-run
  -> would: git push origin main
  -> would: git push --force-with-lease origin <WtBranch>
  -> origin refs unchanged
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupAheadWithOriginBranch(t, req)
	req.Args = []string{"--merge-back", "--push", "--dry-run"}
	return nil
}
```
