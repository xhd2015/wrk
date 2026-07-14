# Scenario

**Feature**: wrk PATH --cd form (path then flag)

```
WRK_FOLLOWUP_FILE set
wrk /WorkRoot/jumpto --cd -> in-place success (same as --cd PATH)
```

## Steps

1. Parent created target at `req.MainRepo`.
2. TargetDir=abs, Args=`["--cd"]`.

```go
func Setup(t *testing.T, req *Request) error {
	setCDPathThenFlag(req, req.MainRepo)
	return nil
}
```
