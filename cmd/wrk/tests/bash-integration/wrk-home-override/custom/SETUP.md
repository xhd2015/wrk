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
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.WrkHome = filepath.Join(req.WorkRoot, "custom-wrk")
	return os.MkdirAll(req.WrkHome, 0o755)
}
```