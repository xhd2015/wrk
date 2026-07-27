# Scenario

**Feature**: WRK_HOME override controls install script path and marker source

```
WRK_HOME=/custom
wrk --bash-integration --install -> script under custom path; marker uses $WRK_HOME
```

## Steps

1. Descendants set custom `req.WrkHome`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "install"
	req.DryRun = false
	return nil
}
```