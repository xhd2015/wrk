# Scenario

**Feature**: in-place wrk --cd /existing/abs writes follow-up cd line

```
WRK_FOLLOWUP_FILE=tmp; workspace/
wrk --cd /WorkRoot/jumpto -> empty stdout; follow-up: cd /WorkRoot/jumpto
```

## Steps

1. Create absolute target directory `{WorkRoot}/jumpto`.
2. Run `wrk --cd <abs>` with follow-up channel open.

```go
func Setup(t *testing.T, req *Request) error {
	target := cdAbsTarget(t, req, "jumpto")
	req.MainRepo = target
	setCDFlagThenPath(req, target)
	return nil
}
```
