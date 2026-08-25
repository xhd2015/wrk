# Scenario

**Feature**: `wrk --bash-integration -h|--help` prints dedicated usage (no action flags)

```
wrk --bash-integration -h|--help -> usage on stdout, exit 0
```

## Steps

- Leaves set help form only (`--help` or `-h`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	return nil
}
```
