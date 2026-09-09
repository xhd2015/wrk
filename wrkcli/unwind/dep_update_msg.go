package unwind

import (
	"sort"
	"strings"
)

// DepUpdateBump is one module version change in a Phase 2 / cascade pin commit.
type DepUpdateBump struct {
	Module string
	From   string
	To     string
}

// FormatDepUpdateCommitMsg builds a one-line subject:
//
//	dep: <module> <from> -> <to>
//	deps: <module> <from> -> <to>, <module> <from> -> <to>
func FormatDepUpdateCommitMsg(bumps []DepUpdateBump) string {
	cp := append([]DepUpdateBump(nil), bumps...)
	sort.Slice(cp, func(i, j int) bool { return cp[i].Module < cp[j].Module })
	var parts []string
	for _, b := range cp {
		if b.Module == "" {
			continue
		}
		parts = append(parts, formatDepUpdatePart(b))
	}
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return "dep: " + parts[0]
	}
	return "deps: " + strings.Join(parts, ", ")
}

func formatDepUpdatePart(b DepUpdateBump) string {
	from := goRequireVersionFromTag(b.From)
	to := goRequireVersionFromTag(b.To)
	switch {
	case from != "" && to != "":
		return b.Module + " " + from + " -> " + to
	case to != "":
		return b.Module + " -> " + to
	case from != "":
		return b.Module + " " + from
	default:
		return b.Module
	}
}
