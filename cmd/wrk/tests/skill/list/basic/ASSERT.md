## Expected Output

```
wrk - <description>
wrk-pipeline - <description>
wrk-pipeline-publish-and-bump - <description>
edit-dependency-code - <description>
```

## Expected

- Exit code 0.
- One line per hosted skill, in registration order.
- Every line is `name - description` with a non-empty description derived from
  that skill's YAML frontmatter.
- No `name/topic` lines (these skills are flat, single-file skills).
- Stderr is empty.

## Side Effects

- Read-only; no install or projects.json mutation.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertSkillListStdout(t, resp.Stdout, resp.Stderr, resp.ExitCode)
}
```
