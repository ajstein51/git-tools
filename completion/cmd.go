package completion

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/astein-peddi/git-tooling/utils"
	"github.com/spf13/cobra"
)

func FetchAllBranches() {
	if !utils.IsInsideGitRepository() {
		return
	}

	cmd := exec.Command("git", "fetch", "--all", "--prune")
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not fetch branches: %v\n", err)
	}
}

func SetupCompletionCommand() *cobra.Command {
	var persist bool

	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate completion scripts",
		Long:  "Generate PowerShell completion scripts for the CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Root()

			if !persist {
				if err := root.GenPowerShellCompletion(os.Stdout); err != nil {
					return fmt.Errorf("failed to generate completion: %w", err)
				}
				
				fmt.Println("\n\033[32mPowerShell completion loaded for this session (not persisted)\033[0m")
				
				return nil
			}

			profilePath := os.ExpandEnv("$PROFILE")
			file, err := os.OpenFile(profilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("failed to open PowerShell profile: %w", err)
			}
			defer file.Close()

			var buf bytes.Buffer
			if err := root.GenPowerShellCompletion(&buf); err != nil {
				return fmt.Errorf("failed to generate completion: %w", err)
			}

			if _, err := file.Write(buf.Bytes()); err != nil {
				return fmt.Errorf("failed to write to PowerShell profile: %w", err)
			}

			fmt.Printf("\n\033[32mPowerShell completion persisted to %s. Restart PowerShell to activate.\033[0m\n", profilePath)
			return nil
		},
	}

	cmd.Flags().BoolVar(&persist, "persist", false, "Persist completion to PowerShell profile")

	return cmd
}
