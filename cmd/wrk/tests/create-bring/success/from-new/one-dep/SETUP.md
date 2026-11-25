# Scenario

**Feature**: `wrk src --new --bring d1` creates + brings (explicit project)

```
src -> wrk <src> --new --no-config --bring <d1>
  -> new default WT; bring into that WT
  -> event command=create; args include --new, --bring, and d1
# bare `wrk --new --bring <d1>` (no project) is rejected: create+bring needs <project>
```

## Steps

1. Create `src` requiring dep1 + `mydep1`.
2. Run with `src` as the project positional (L2).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	src := initCreateBringSrc(t, req.WorkRoot, createBringDep1Module)
	dep1 := initCreateBringDep(t, req.WorkRoot, "mydep1", createBringDep1Module)

	req.TargetDir = src
	req.RepoDir = src
	req.MainRepo = src
	req.DepPath = dep1
	req.ConsumerTop = createBringDefaultWT(req)
	req.Args = []string{"--new", "--no-config", "--bring", dep1}
	return nil
}
```
