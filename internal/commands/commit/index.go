package commit

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/Lachine1/claude-gode/pkg/types"
)

// New creates the /commit command.
func New(isGit bool, gitRoot string) types.Command {
	return types.Command{
		Name:        "commit",
		Aliases:     []string{},
		Description: "Commit changes with git",
		Usage:       "/commit [message]",
		Handler: func(ctx *types.CommandContext, args []string) error {
			return handleCommit(ctx, args, isGit, gitRoot)
		},
	}
}

func handleCommit(ctx *types.CommandContext, args []string, isGit bool, gitRoot string) error {
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

	if len(strings.TrimSpace(string(statusOutput))) == 0 {
		fmt.Println()
		fmt.Println("  Nothing to commit. Working tree is clean.")
		fmt.Println()
		return nil
	}

	stageCmd := exec.Command("git", "-C", cwd, "add", "-A")
	if err := stageCmd.Run(); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	var message string
	if len(args) > 0 {
		message = strings.Join(args, " ")
	} else {
		diffCmd := exec.Command("git", "-C", cwd, "diff", "--cached", "--stat")
		diffOutput, err := diffCmd.Output()
		if err != nil || len(diffOutput) == 0 {
			message = "chore: update files"
		} else {
			message = fmt.Sprintf("chore: %s", strings.TrimSpace(string(diffOutput)))
			if len(message) > 72 {
				message = message[:72]
			}
		}
	}

	commitCmd := exec.Command("git", "-C", cwd, "commit", "-m", message)
	commitOutput, err := commitCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("commit failed: %s\n%s", err, string(commitOutput))
	}

	fmt.Printf("\n  Committed: %s\n", message)
	fmt.Printf("  %s\n", strings.TrimSpace(string(commitOutput)))
	fmt.Println()

	return nil
}
