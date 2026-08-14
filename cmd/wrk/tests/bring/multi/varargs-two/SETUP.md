# Scenario

**Feature**: two required deps via one `Varargs` `--bring` (`wrk --bring p1 p2`)

```
# same side effects as two-success; one flag, two values
consumer requires dep1+dep2
  -> wrk --bring <dep1> <dep2>
  -> two external abs paths on stdout (order = bring args)
  -> replace for both; /external gitignore; command=bring; args include --bring and both paths
```

## Steps

1. Create consumer requiring `example.com/dep1` and `example.com/dep2`.
2. Create dep repos `mydep1` / `mydep2`.
3. Run `wrk --bring <dep1> <dep2>` (one `--bring`, two values) from consumer (L2).

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
	req.Args = []string{"--bring", dep1, dep2}
	return nil
}
```
