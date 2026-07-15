## Expected

- `err` is nil.
- `Primary` is `[main, inTree, linkedLate, linkedEarly]`:
  - main first
  - then ListLinked porcelain order (not path sort of out-of-tree pair)
  - in-tree linked once only (also present in scan)
- `External` is only nested/dep paths, path-sorted:
  `[external/child, task-hub, tools/child]`
- In-tree linked and out-of-tree linked must **not** appear in external.

## Side Effects

- None (pure helper).

## Exit Code

- N/A (no process).

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPartition(t, req, resp, err)
}
```
