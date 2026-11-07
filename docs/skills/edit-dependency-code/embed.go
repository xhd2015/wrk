// Package editdependencycode embeds the edit-dependency-code agent skill
// (docs/skills/edit-dependency-code/SKILL.md).
package editdependencycode

import _ "embed"

//go:embed SKILL.md
var SkillContent string
