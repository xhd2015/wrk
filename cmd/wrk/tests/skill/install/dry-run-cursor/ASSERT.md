## Expected Output

```
[dry-run] Installed skill to: <workdir>/.cursor/skills/wrk
[dry-run]   create: <workdir>/.cursor/skills/wrk/SKILL.md
```

## Expected

- Exit code 0.
- Stdout lists the planned `.cursor/skills/wrk` directory and `SKILL.md` file.
- `{RepoDir}/.cursor/skills/wrk` does NOT exist after the run.
- Stderr is empty.

## Side Effects

- Dry-run only; no files or directories created.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertOutputExact(t, resp.Stdout, installDryRunCursorStdoutV2(t, req.RepoDir))
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertFileNotExists(t, cursorSkillInstallDir(req.RepoDir))
}
```
