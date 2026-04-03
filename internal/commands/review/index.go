package review

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// New creates the /review command.
func New(isGit bool, gitRoot string) types.Command {
	return types.Command{
		Name:        "review",
		Aliases:     []string{"diff"},
		Description: "Review changes",
		Usage:       "/review [file]",
		Handler: func(ctx *types.CommandContext, args []string) error {
			return handleReview(ctx, args, isGit, gitRoot)
		},
	}
}

func handleReview(ctx *types.CommandContext, args []string, isGit bool, gitRoot string) error {
	if !isGit {
		return fmt.Errorf("not in a git repository")
	}

	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is not installed")
	}

	cwd := gitRoot
	if cwd == "" {
		cwd = ctx.Cwd
	}

	statusCmd := exec.Command("git", "-C", cwd, "status", "--short")
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return fmt.Errorf("git error: %w", err)
	}

	if len(statusOutput) == 0 {
		ctx.WriteOutput("")
		ctx.WriteOutput("  No changes to review. Working tree is clean.")
		ctx.WriteOutput("")
		return nil
	}

	w := ctx.WriteOutput
	w("")
	w("  Review Changes")
	w("  ═══════════════════════════════════════")
	w("")
	w("  Changed files:")
	w("")

	lines := strings.Split(strings.TrimSpace(string(statusOutput)), "\n")
	for _, line := range lines {
		if len(strings.TrimSpace(line)) > 0 {
			w("    " + strings.TrimSpace(line))
		}
	}

	w("")
	w(fmt.Sprintf("  Total: %d changed file(s)", len(lines)))
	w("")

	if len(args) > 0 {
		fileName := args[0]
		diffCmd := exec.Command("git", "-C", cwd, "diff", "--", fileName)
		diffOutput, err := diffCmd.Output()
		if err == nil && len(diffOutput) > 0 {
			w("  Diff for " + fileName + ":")
			w("")
			diffLines := strings.Split(string(diffOutput), "\n")
			maxLines := 50
			if len(diffLines) > maxLines {
				diffLines = diffLines[:maxLines]
				w("    (showing first 50 lines)")
				w("")
			}
			for _, line := range diffLines {
				w("    " + line)
			}
			w("")
		}
	}

	return nil
}
