package registry

// Agent skills are portable workflow guides (agentskills.io format) served by the
// backend over MCP. They hold instructions only, never tenant data.
func init() {
	Register(&Domain{
		Name:        "agent-skill",
		Aliases:     []string{"agent-skills"},
		Description: "List and read UteamUP workflow skills for AI agents",
		Actions: []Action{
			{Name: "list", Description: "List the workflow skills visible to this connection", ToolName: "UteamupAgentSkillList", MCPOnly: true, RejectExtraArgs: true},
			{Name: "get", Description: "Read one skill file (SKILL.md by default)", ToolName: "UteamupAgentSkillGet", MCPOnly: true, RejectExtraArgs: true,
				Args:  []ArgDef{{Name: "name", Description: "Skill name, for example uteamup-workorder-triage", Required: true, Type: "string"}},
				Flags: []FlagDef{{Name: "file", Description: "File inside the skill, for example references/intervals.md", Type: "string"}}},
		},
	})
}
