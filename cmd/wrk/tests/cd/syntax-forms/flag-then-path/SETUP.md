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
func Setup(t *testing.T, req *Request) error {
	setCDFlagThenPath(req, req.MainRepo)
	return nil
}
```
