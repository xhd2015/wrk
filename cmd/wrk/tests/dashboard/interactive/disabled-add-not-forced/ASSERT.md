## Expected

- If exit **0**: compose argv log present for DONE recipe with **`--dry-run`**, **without** `--add-all`.
- If exit **non-zero**: must not have applied compose with `--add-all`; argv log if written must still omit `--add-all`; prefer message mentioning gate / disabled / add.
- Clean linked worktree still present.

## Side Effects

- No force-on of disabled Add changes into recipe.

## Exit Code

- 0 preferred (ignore illegal toggle); non-zero gate error also acceptable if no `--add-all`.

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertLinkedWorktreeStillPresent(t, req)

	raw := strings.TrimSpace(readComposeArgvLog(t, req))
	if resp.ExitCode == 0 {
		if raw == "" {
			t.Fatalf("exit 0 RUN must write compose argv log; path=%s", dashComposeArgvLogPath(req))
		}
		assertComposeArgvRecipeDone(t, req, true /* dryRun */, false /* no add-all */)
		assertDryRunComposeEvidence(t, resp, "done")
		return
	}
	// Non-zero: gate rejection path — still must not cook --add-all.
	if raw != "" {
		toks := composeArgvTokens(raw)
		if argvHasToken(toks, "--add-all") {
			t.Fatalf("disabled Add changes must not appear as --add-all; tokens=%v", toks)
		}
	}
	// Soft: stderr/stdout mention disable/gate/add
	all := strings.ToLower(resp.Stdout + "\n" + resp.Stderr)
	if !strings.Contains(all, "add") && !strings.Contains(all, "disabled") &&
		!strings.Contains(all, "gate") && !strings.Contains(all, "cannot") {
		// still OK if argv empty and no compose ran
		if raw != "" {
			t.Fatalf("non-zero without clear gate signal and argv=%q out=%q err=%q",
				raw, resp.Stdout, resp.Stderr)
		}
	}
}
```
