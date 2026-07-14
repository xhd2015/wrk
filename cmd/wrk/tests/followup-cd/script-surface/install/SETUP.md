# Scenario

**Feature**: install writes integration/bash.sh with follow-up wrapper

```
wrk --bash-integration --install -> {WRK_HOME}/integration/bash.sh
```

## Steps

1. Set Mode to install.

```go
func Setup(t *testing.T, req *Request) error {
	req.Mode = "install"
	return nil
}
```
