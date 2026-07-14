# Scenario

**Feature**: non-create modes skip basename fallback (except --list, --status, and --repos)

```
# saved project exists but mode is not create/list/status -> no lookup
wrk <basename> --done -> normal missing-dir error
```

## Steps

- Descendants record a saved project and invoke a non-create mode with basename `<dir>`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureBasenameFallbackHelpersUsed()
	return nil
}
```