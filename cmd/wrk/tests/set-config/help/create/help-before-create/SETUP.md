# Scenario

**Feature**: help flag order — `--help` before `--create` still yields create-level help

```
workspace/ -> wrk --set-config --help --create
  -> dedicated create usage (dispatch rule: help + create → create level)
  -> exit 0
```

## Steps

1. Run `wrk --set-config --help --create` (help token before action).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = setConfigArgs("--help", "--create")
	return nil
}
```
