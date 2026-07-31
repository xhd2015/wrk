# Scenario

**Feature**: two required deps via repeatable `--bring` create two external worktrees + replaces

```
# consumer requires dep1+dep2
# wrk --bring <dep1> --bring <dep2>
#   -> two external abs paths on stdout (order = bring args)
#   -> replace for both modules; /external gitignore; exit 0
consumer + mydep1 + mydep2
  -> wrk --bring mydep1 --bring mydep2 -> success
```

## Steps

1. Create consumer requiring `example.com/dep1` and `example.com/dep2`.
2. Create dep repos `mydep1` / `mydep2` with those modules.
3. Run `wrk --bring <dep1> --bring <dep2>` from consumer (L2 InProcess).

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
	req.Args = []string{"--bring", dep1, "--bring", dep2}
	return nil
}
```
