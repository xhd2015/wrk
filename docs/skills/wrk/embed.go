// Package wrkskill embeds the wrk agent skill (docs/skills/wrk/SKILL.md).
package wrkskill

import _ "embed"

//go:embed SKILL.md
var SkillContent string
