# Scenario

**Feature**: root help documents done/merge-back composition with post modifiers

```
# user-facing usage() matches composition: primary + optional sync/tag-next/push/dry-run
wrk --help
  -> exit 0
  -> --done / --merge-back synopsis lists optional post modifiers
  -> --push dual meaning (not only with --tag-next)
  -> no claim that --tag-next is exclusive with --done
```

## Steps

1. Run `wrk --help` from isolated WorkRoot (no git required).

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--help"}
	return nil
}
```
