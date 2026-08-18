# Scenario

**Feature**: other-checkout module with an existing replace (no require) is still gated

```
# kool has replace for dep but no require
cwd=primary -> wrk --dep-replace <dep>
  -> rewrite kool's replace to the new absolute NewPath
  -> primary (requires dep) also rewritten
```

## Steps

1. Seed stack where kool already replaces dep (no require).
2. Run apply from primary.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupStackExistingReplace(t, req)
	req.Args = []string{"--dep-replace", req.DepDir}
	return nil
}
```
