# Scenario

**Feature**: `wrk --new --bring d1` from inside `src` creates + brings

```
src cwd -> wrk --new --no-config --bring <d1>
  -> new default WT; bring into that WT
  -> event command=create; args include --new, --bring, and d1
```

## Steps

1. Create `src` requiring dep1 + `mydep1`.
2. Run from `src` (L2).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	src := initCreateBringSrc(t, req.WorkRoot, createBringDep1Module)
	dep1 := initCreateBringDep(t, req.WorkRoot, "mydep1", createBringDep1Module)

	req.RepoDir = src
	req.MainRepo = src
	req.DepPath = dep1
	req.ConsumerTop = createBringDefaultWT(req)
	req.Args = []string{"--new", "--no-config", "--bring", dep1}
	return nil
}
```
