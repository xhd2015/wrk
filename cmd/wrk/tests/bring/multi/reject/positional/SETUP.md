# Scenario

**Feature**: leftover positionals with `--bring` are rejected (no multi-value sugar)

```
# wrk --bring p1 p2  (one flag + extra positional)
#   -> non-zero; unexpected arguments
#   -> today leftover positional can be misread as consumer workDir — must not
consumer + mydep1 present
  -> wrk --bring <dep1> <dep2-as-positional>
```

## Steps

1. Create consumer + two valid dep paths (so multi-flag form would work).
2. Run `wrk --bring <dep1> <dep2>` with **one** `--bring` and dep2 as a bare positional.

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
	// One --bring + extra positional (not a second --bring flag).
	req.Args = []string{"--bring", dep1, dep2}
	return nil
}
```
