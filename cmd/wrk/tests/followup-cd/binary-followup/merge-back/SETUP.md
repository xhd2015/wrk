# Scenario

**Feature**: --merge-back does not write follow-up cd

```
linked wt + WRK_FOLLOWUP_FILE
wrk --merge-back -> success possible; follow-up empty
```

## Steps

1. Descendants set --merge-back args.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	requireMode(t, req, "binary")
	return nil
}
```
