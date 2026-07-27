# Scenario

**Feature**: bash-integration modes are mutually exclusive with normal wrk commands

```
wrk --bash-integration combined with other modes -> error
```

## Steps

1. Set `req.Mode = "mutual"`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "mutual"
	req.CLIArgs = []string{"--bash-integration", "--list"}
	return nil
}
```