# Scenario

**Feature**: failed CPR does not establish known origin (dual-origin path)

```
# empty / incomplete read after timeout
buf = []
  -> ParseCPR → !ok
  -> OriginOK false (no known origin for ResolveMouseHit)
  # live TUI would leave OriginY nil → dual-origin (P1)
```

## Steps

1. Inject empty buffer so parse fails; do not synthesize CPR.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Empty Buf + BlankAbove < 0 disables synthetic CPR in Run (fail path).
	req.Buf = []byte{}
	req.BlankAbove = -1
	return nil
}
```
