# Scenario

**Feature**: successful empty plan when nothing is a reinstall candidate

```
# go.mod ok but no package main under cmd or script/.../install
PlanLocalReinstalls
  -> Items=[] (ok, not an error)
```

## Preconditions

- Module has valid go.mod; discovery finds zero candidates.

## Steps

1. Leaves write go.mod only (or non-candidate packages only, if documented).
2. Assert empty Items and nil error.

## Context

- Empty plan is success — later CLI phases may print "nothing to do".

```go
func Setup(t *testing.T, req *Request) error {
	req.WantItems = []WantPlanItem{}
	return nil
}
```
