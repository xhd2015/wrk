# Scenario

**Feature**: fallback propagates non-zero interactive shell exit code

```
channel closed; fake bash exits 42
wrk --cd /jumpto -> wrk exit code 42; stdout still abs path; install hint
```

## Steps

1. Create abs target; install fake bash with exit 42.
2. Run `wrk --cd <abs>`.

```go
func Setup(t *testing.T, req *Request) error {
	target := cdAbsTarget(t, req, "jumpto")
	req.MainRepo = target
	installFakeBash(t, req, 42)
	setCDFlagThenPath(req, target)
	return nil
}
```
