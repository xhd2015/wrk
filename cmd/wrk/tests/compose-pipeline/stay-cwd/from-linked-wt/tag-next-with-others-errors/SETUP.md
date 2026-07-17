# Scenario

**Feature**: any compose including `--tag-next` from linked wt without done/merge-back errors

```
linked wt -> wrk --sync --tag-next --push
  -> non-zero
  -> names --tag-next + main requirement
  -> does not partially apply push/tag on main under the wrong activeRoot
```

## Steps

1. Linked ahead + origin.
2. Run multi-stage including tag-next without done.

```go
func Setup(t *testing.T, req *Request) error {
	setupAPLinkedAheadOrigin(t, req)
	req.Args = []string{"--sync", "--tag-next", "--push"}
	return nil
}
```
