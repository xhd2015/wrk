# Scenario

**Feature**: `wrk <url> --where --pr` same as flag-first form

```
recorded linked + fake gh
  -> wrk https://github.com/acme/app/pull/42 --where --pr
  -> stdout linked path
```

## Steps

1. Seed recorded linked fixture.
2. Place URL as first positional (`TargetDir`); flags `--where --pr` in Args.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InProcess = true
	wherePrSetupRecordedLinked(t, req)
	setWherePrURLFirst(req, wherePrURL)
	return nil
}
```
