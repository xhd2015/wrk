# Scenario

**Feature**: Status returns a structured checkout report

```
# checkout path
Caller -> workops.Status(checkout) -> StatusReport
```

## Steps

1. Grouping only: set Op to status.
2. Leaves seed fixtures and Checkout.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = OpStatus
	return nil
}
```
