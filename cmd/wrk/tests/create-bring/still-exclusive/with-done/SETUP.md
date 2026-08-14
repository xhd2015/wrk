# Scenario

**Feature**: `wrk --done --bring d1` stays mutually exclusive

```
# compose is create+bring only; --done is still another mode
src (git) -> wrk --done --bring <d1>
  -> non-zero; mutually exclusive
  -> no new WT; no external/
```

## Steps

1. Create `src` + a valid dep (so compose would otherwise be possible).
2. Run `--done --bring d1` from `src`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	src := initCreateBringSrc(t, req.WorkRoot, createBringDep1Module)
	dep1 := initCreateBringDep(t, req.WorkRoot, "mydep1", createBringDep1Module)
	req.RepoDir = src
	req.MainRepo = src
	req.DepPath = dep1
	req.Args = []string{"--done", "--bring", dep1}
	return nil
}
```
