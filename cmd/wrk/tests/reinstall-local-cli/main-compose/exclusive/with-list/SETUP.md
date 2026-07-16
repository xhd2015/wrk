# Scenario

**Feature**: --main --reinstall-local --list is mutually exclusive (MC2)

```
# MC2: compose + --list still exclusive
workspace/ -> wrk --main --reinstall-local --list
  -> non-zero; mutually exclusive; empty stdout
```

## Steps

1. Run `wrk --main --reinstall-local --list` from neutral module dir (no go.mod
   required; exclusion should fire before planning / shell).

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--main", "--reinstall-local", "--list"}
	return nil
}
```
