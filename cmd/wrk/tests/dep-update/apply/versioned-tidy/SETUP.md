# Scenario

**Feature**: dir-mode tidy uses the consumer go line pin (go 1.19 → go1.19.13)

```
# consumer go.mod has go 1.19; wrapper at $InstallDir/go1.19.13/bin/go
nearest consumer + tagged dep
  -> wrk --dep-update <dep>
  -> pin + go mod tidy ok
  -> wrapper last-run records GOROOT/PATH0 suffix go1.19.13
```

## Steps

1. Seed consumer with `go 1.19`, file:// GOPROXY, and `go1.19.13` host-go wrapper.
2. Run apply.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupVersionedTidy(t, req)
	req.Args = []string{"--dep-update", req.DepDir, "-v"}
	return nil
}
```
