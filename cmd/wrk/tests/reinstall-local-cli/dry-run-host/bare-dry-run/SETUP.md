# Scenario

**Feature**: bare wrk --dry-run is rejected and lists --reinstall-local as a host

```
# C5
workspace/ -> wrk --dry-run
  -> non-zero
  -> stderr mentions --dry-run is only valid with …
  -> stderr mentions --reinstall-local
```

## Steps

1. Run `wrk --dry-run` alone from the empty module dir (no host flags).

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--dry-run"}
	return nil
}
```
