## Expected

- `err` is nil.
- `resp.Projects` contains an entry whose Path matches the registered main
  (normalized absolute path).

## Side Effects

- None (read-only list).

## Errors

- None.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	want := resolvePath(t, req.MainRepo)
	found := false
	for _, p := range resp.Projects {
		if resolvePath(t, p.Path) == want {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ListProjects missing %q; got %+v", want, resp.Projects)
	}
}
```
