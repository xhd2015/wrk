# Scenario

**Feature**: shared two-linked branch refuses manual commit

```
wt1 (feature-shared, staged) + wt2 (feature-shared)
  -> wrk --commit -m "feat: refuse"
  -> non-zero; Error: refuse commit
  -> HEAD subject unchanged; change.go stays staged
```

## Steps

1. Build shared two-linked staged fixture.
2. Snapshot HEAD subject.
3. Run `wrk --commit -m "feat: refuse"` from primary wt.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSharedTwoLinkedStaged(t, req)
	req.HEADSubject = gitHEADSubject(t, req.RepoDir)
	req.Args = []string{"--commit", "-m", "feat: refuse"}
	return nil
}
```
