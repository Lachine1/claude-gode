package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// New creates the /skills command.
func New() types.Command {
	return types.Command{
		Name:        "skills",
		Aliases:     []string{},
		Description: "List available skills",
		Usage:       "/skills [skill-name]",
		Handler: func(ctx *types.CommandContext, args []string) error {
			return handleSkills(ctx, args)
		},
	}
}

func handleSkills(ctx *types.CommandContext, args []string) error {
	skillDirs := findSkillDirectories(ctx.Cwd)

	if len(args) > 0 {
		return showSkillDetail(ctx, args[0], skillDirs)
	}
	return listSkills(ctx, skillDirs)
}

func listSkills(ctx *types.CommandContext, skillDirs []string) error {
	w := ctx.WriteOutput
	w("")
	w("  Available Skills")
	w("  ═══════════════════════════════════════")
	w("")

	if len(skillDirs) == 0 {
		w("  No skills found.")
		w("  Skills are defined in AGENTS.md or .claude/skills/ directory.")
	} else {
		for _, dir := range skillDirs {
			name := filepath.Base(dir)
			desc := readSkillDescription(dir)
			w(fmt.Sprintf("  %-25s %s", name, desc))
		}
	}

	w("")
	w(fmt.Sprintf("  Total: %d skill(s)", len(skillDirs)))
	w("")
	w("  Use /skills <skill-name> for more details.")
	w("")
	return nil
}

func showSkillDetail(ctx *types.CommandContext, name string, skillDirs []string) error {
	for _, dir := range skillDirs {
		if filepath.Base(dir) == name {
			desc := readSkillDescription(dir)
			w := ctx.WriteOutput
			w("")
			w("  Skill: " + name)
			w("  ═══════════════════════════════════════")
			w("")
			w("  Description: " + desc)
			w("")

			entries, err := os.ReadDir(dir)
			if err == nil {
				w("  Files:")
				for _, entry := range entries {
					if !entry.IsDir() {
						w("    - " + entry.Name())
					}
				}
			}

			w("")
			return nil
		}
	}

	return fmt.Errorf("unknown skill: %s", name)
}

func findSkillDirectories(cwd string) []string {
	var skills []string

	skillsDir := filepath.Join(cwd, ".claude", "skills")
	if entries, err := os.ReadDir(skillsDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				skills = append(skills, filepath.Join(skillsDir, entry.Name()))
			}
		}
	}

	agentsPath := filepath.Join(cwd, "AGENTS.md")
	if _, err := os.Stat(agentsPath); err == nil {
		skills = append(skills, agentsPath)
	}

	return skills
}

func readSkillDescription(dir string) string {
	if filepath.Base(dir) == "AGENTS.md" {
		return "Agent instructions file"
	}

	skillPath := filepath.Join(dir, "SKILL.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		return "No description"
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			if len(line) > 80 {
				return line[:77] + "..."
			}
			return line
		}
	}
	return "No description"
}
