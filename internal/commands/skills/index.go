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
		return showSkillDetail(args[0], skillDirs)
	}
	return listSkills(skillDirs)
}

func listSkills(skillDirs []string) error {
	fmt.Println()
	fmt.Println("  Available Skills")
	fmt.Println("  ═══════════════════════════════════════")
	fmt.Println()

	if len(skillDirs) == 0 {
		fmt.Println("  No skills found.")
		fmt.Println("  Skills are defined in AGENTS.md or .claude/skills/ directory.")
	} else {
		for _, dir := range skillDirs {
			name := filepath.Base(dir)
			desc := readSkillDescription(dir)
			fmt.Printf("  %-25s %s\n", name, desc)
		}
	}

	fmt.Printf("\n  Total: %d skill(s)\n", len(skillDirs))
	fmt.Println()
	fmt.Println("  Use /skills <skill-name> for more details.")
	fmt.Println()
	return nil
}

func showSkillDetail(name string, skillDirs []string) error {
	for _, dir := range skillDirs {
		if filepath.Base(dir) == name {
			desc := readSkillDescription(dir)
			fmt.Println()
			fmt.Printf("  Skill: %s\n", name)
			fmt.Println("  ═══════════════════════════════════════")
			fmt.Println()
			fmt.Printf("  Description: %s\n\n", desc)

			entries, err := os.ReadDir(dir)
			if err == nil {
				fmt.Println("  Files:")
				for _, entry := range entries {
					if !entry.IsDir() {
						fmt.Printf("    - %s\n", entry.Name())
					}
				}
			}

			fmt.Println()
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
