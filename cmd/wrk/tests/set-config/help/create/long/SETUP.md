# Scenario

**Feature**: `wrk --set-config --create --help` prints dedicated create usage

```
workspace/ -> wrk --set-config --create --help
  -> create UX usage on stdout, exit 0 (user repro path)
```

## Steps

1. Run `wrk --set-config --create --help` (no UX write flags).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = setConfigArgs("--create", "--help")
	return nil
}
```
