# Scenario

**Feature**: wrk --cd PATH form (flag then path)

```
WRK_FOLLOWUP_FILE set
wrk --cd /WorkRoot/jumpto -> in-place success
```

## Steps

1. Parent created target at `req.MainRepo`.
2. Args = `["--cd", abs]`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setCDFlagThenPath(req, req.MainRepo)
	return nil
}
```
