# Scenario

**Feature**: `--where --pr URL --status` is mutually exclusive

```
recorded linked + fake gh
  -> wrk --where --pr URL --status
  -> non-zero; empty stdout
  -> stderr mutually exclusive / mode conflict
```

## Steps

1. Seed recorded linked + fake gh (so failure is mutex, not missing fixture).
2. Args include `--status` alongside compose.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InProcess = true
	wherePrSetupRecordedLinked(t, req)
	req.Args = []string{"--where", "--pr", wherePrURL, "--status"}
	return nil
}
```
