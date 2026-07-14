# Scenario

**Feature**: status reports partial when bash.sh exists without profile markers

```
bash.sh present, no markers in profiles
wrk --bash-integration --status -> partial, exit 1
```

## Steps

1. Pre-seed bash.sh only.

```go
func Setup(t *testing.T, req *Request) error {
	req.PreExistingBashSh = minimalBashSh()
	return nil
}
```