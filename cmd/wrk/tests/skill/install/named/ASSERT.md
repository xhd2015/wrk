## Expected Output

```
[dry-run] Installed skill to: <workdir>/.cursor/skills/wrk-pipeline-publish-and-bump
[dry-run]   create: <workdir>/.cursor/skills/wrk-pipeline-publish-and-bump/SKILL.md
```

## Expected

- Exit code 0.
- The install directory is named after the requested skill, not the default.
- Nothing is written (dry-run).
- Stderr is empty.

## Side Effects

- Dry-run only; no files or directories created.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	const name = "wrk-pipeline-publish-and-bump"
	assertOutputExact(t, resp.Stdout, installDryRunCursorStdoutV2(t, req.RepoDir, name))
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertFileNotExists(t, cursorSkillInstallDir(req.RepoDir, name))
}
```
