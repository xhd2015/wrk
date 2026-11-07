// Package wrkpipeline embeds the wrk-pipeline agent skill (docs/skills/wrk-pipeline/SKILL.md).
package wrkpipeline

import _ "embed"

//go:embed SKILL.md
var SkillContent string
