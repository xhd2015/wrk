# Scenario

**Feature**: custom WRK_HOME install writes script under override path

```
WRK_HOME={WorkRoot}/custom-wrk
wrk --bash-integration --install -> {custom}/integration/bash.sh + marker sources WRK_HOME
```

## Steps

1. Set `req.WrkHome` to `{WorkRoot}/custom-wrk`.
2. Run install.

```go
import (
	"os"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	req.WrkHome = filepath.Join(req.WorkRoot, "custom-wrk")
	return os.MkdirAll(req.WrkHome, 0o755)
}
```