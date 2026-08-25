# Scenario

**Feature**: help flag order — `--help` before `--install` still yields usage (no write)

```
workspace/ -> wrk --bash-integration --help --install
  -> usage on stdout, exit 0; no filesystem writes
```

## Steps

1. Run `wrk --bash-integration --help --install` (help token before action).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.CLIArgs = []string{"--bash-integration", "--help", "--install"}
	return nil
}
```
