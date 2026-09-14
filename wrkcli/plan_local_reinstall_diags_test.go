package wrkcli

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintReinstallDiagnosticsToUsesErrW(t *testing.T) {
	var errBuf bytes.Buffer
	diags := []ReinstallDiagnostic{{
		Level:   DiagLevelNotice,
		Kind:    DiagKindPreferScript,
		BinName: "wrk",
		Paths:   []string{"./cmd/wrk", "./script/wrk"},
	}}
	printReinstallDiagnosticsTo(&errBuf, diags, false)
	got := errBuf.String()
	if !strings.Contains(got, "notice: bin wrk:") {
		t.Fatalf("expected notice on errW; got %q", got)
	}
	if !strings.Contains(got, "preferring") {
		t.Fatalf("expected prefer-script body; got %q", got)
	}
}

func TestExecuteMultiLocalReinstallsToRoutesDiagnosticsToErrW(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	plan := &MultiLocalReinstallPlan{
		BinDir: t.TempDir(),
		Modules: []ModuleReinstallPlan{{
			ModuleRoot: t.TempDir(),
			ModulePath: "example.com/m",
			ModuleName: "m",
			RelDir:     ".",
			Items:      nil, // no installs; only diagnostics
			Diagnostics: []ReinstallDiagnostic{{
				Level:   DiagLevelNotice,
				Kind:    DiagKindPreferScript,
				BinName: "tool",
				Paths:   []string{"./cmd/tool", "./script/tool"},
			}},
		}},
	}
	st, err := executeMultiLocalReinstallsTo(plan, false, false, &outBuf, &errBuf)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if st.Reinstalled != 0 || st.Skipped != 0 || st.Failed != 0 {
		t.Fatalf("stats=%+v", st)
	}
	if !strings.Contains(errBuf.String(), "notice: bin tool:") {
		t.Fatalf("diagnostics must go to errW; err=%q out=%q", errBuf.String(), outBuf.String())
	}
	if strings.Contains(outBuf.String(), "notice:") {
		t.Fatalf("diagnostics must not go to out; out=%q", outBuf.String())
	}
	if !strings.Contains(outBuf.String(), "reinstalled 0") {
		t.Fatalf("expected summary on out; out=%q", outBuf.String())
	}
}
