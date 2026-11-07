// Package wrkpipelinepublishandbump embeds the wrk-pipeline-publish-and-bump
// agent skill (docs/skills/wrk-pipeline-publish-and-bump/SKILL.md).
package wrkpipelinepublishandbump

import _ "embed"

//go:embed SKILL.md
var SkillContent string
