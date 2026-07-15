# Scenario

**Feature**: wrk --tag-next rejects other mode flags

```
# --tag-next combined with --done/--list/etc -> mutually exclusive error
wrk --tag-next --done -> error
```

## Steps

- Descendants combine `--tag-next` with another standalone mode flag.

```go
func Setup(t *testing.T, req *Request) error {
	tagNextEnsureHelpersUsed()
	return nil
}
```