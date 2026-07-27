# Scenario

**Feature**: wrk --scan-git-repos --projects is mutually exclusive

```
wrk --scan-git-repos --projects -> non-zero; mutually exclusive; empty stdout
```

## Steps

1. Run `wrk --scan-git-repos --projects` from isolated WorkRoot.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--scan-git-repos", "--projects"}
	return nil
}
```
