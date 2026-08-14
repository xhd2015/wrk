# Scenario

**Feature**: create + single `--bring` + `--exec pwd` runs pwd in the **project** WT

```
# exclusive multi+exec reject does not apply (one bring path)
src cwd -> wrk --new --no-config --bring <d1> --exec pwd
  -> stdout contains create path, external path, and pwd == project WT (not external)
```

## Steps

1. Create `src` requiring dep1 + `mydep1`.
2. Run `--new --bring d1 --exec pwd`.

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
	req.Args = []string{"--new", "--no-config", "--bring", dep1, "--exec", "pwd"}
	return nil
}
```
