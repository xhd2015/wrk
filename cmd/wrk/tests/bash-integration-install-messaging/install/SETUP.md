# Scenario

**Feature**: real install writes assets and prints status report

```
# no --dry-run
wrk --bash-integration --install
  -> may write bash.sh / append markers
  -> stdout: bash integration: installed|updated|is up to date
```

## Steps

1. Set `req.Mode = "install"` and `req.DryRun = false`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "install"
	req.DryRun = false
	return nil
}
```
