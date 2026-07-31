# Scenario

**Feature**: `--exec` is rejected when more than one `--bring` path is given

```
consumer + two deps
  -> wrk --bring d1 --bring d2 --exec pwd
  -> non-zero; clear error about --exec + single --bring
  -> no external worktrees created (reject before loop preferred)
```

## Steps

1. Create consumer requiring both deps + both dep repos (so multi without exec would succeed).
2. Run multi-bring with `--exec pwd`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initMultiBringConsumerWithTwoRequires(t, req.WorkRoot)
	dep1 := initMultiBringDepRepo(t, req.WorkRoot, "mydep1", multiBringDep1Module)
	dep2 := initMultiBringDepRepo(t, req.WorkRoot, "mydep2", multiBringDep2Module)

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.DepPath = dep1
	req.SecondRepo = dep2
	req.Args = []string{"--bring", dep1, "--bring", dep2, "--exec", "pwd"}
	return nil
}
```
