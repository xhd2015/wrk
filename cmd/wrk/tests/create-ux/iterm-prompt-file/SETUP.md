# Scenario

**Feature**: long `-t` agent prompt uses agent-pro `--prompt-file` spill

```
wrk -t '<prompt longer than 600 runes>' --new-terminal --open-in-agent
  -> iTerm write-text follow-up has --prompt-file=<abs>
  -> spill file body is /brainstorm <full task>
  -> long body is not embedded in AppleScript

wrk -t '<same>' --open-in-agent
  -> in-process agent-run argv has --prompt-file; last token is not the body
```

## Steps

- Leaves set a TaskDesc whose `/brainstorm ${task}` exceeds
  `agentrunapi.PromptFileSpillMinRunes` (600).
- Reuse create-ux mocks (darwin hermetic).

```go
import (
	"strings"
	"unicode/utf8"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	setupMainRepoForCreateUX(t, req)
	installCreateUXMocks(t, req, "darwin")
	// 700 ASCII runes plus "/brainstorm " is well over PromptFileSpillMinRunes.
	req.TaskDesc = strings.Repeat("x", 700)
	req.TaskFlag = "-t"
	n := utf8.RuneCountInString("/brainstorm " + req.TaskDesc)
	if n <= agentrunapi.PromptFileSpillMinRunes {
		t.Fatalf("fixture prompt must exceed %d runes; got %d",
			agentrunapi.PromptFileSpillMinRunes, n)
	}
	return nil
}
```
