# Scenario

**Feature**: `wrk --pr --where URL` same as `--where --pr URL`

```
recorded linked + fake gh
  -> wrk --pr --where https://github.com/acme/app/pull/42
  -> stdout linked path
```

## Steps

1. Seed recorded linked fixture.
2. Args: `--pr --where <url>`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	wherePrSetupRecordedLinked(t, req)
	req.Args = wherePrThenWhereArgs(wherePrURL)
	return nil
}
```
