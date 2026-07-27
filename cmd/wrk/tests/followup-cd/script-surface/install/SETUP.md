# Scenario

**Feature**: install writes integration/bash.sh with follow-up wrapper

```
wrk --bash-integration --install -> {WRK_HOME}/integration/bash.sh
```

## Steps

1. Set Mode to install.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "install"
	return nil
}
```
