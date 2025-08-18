package prs

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/astein-peddi/git-tooling/utils"
	"github.com/spf13/cobra"
)

// move the getBranchNames to a function here
func GetBranchNames() []string {

	if !utils.IsInsideGitRepository() {
		return []string{}
	}

	cmd := exec.Command("git", "branch", "--all", "--format=%(refname:short)")
	out, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var branches []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && line != "HEAD" {
			branches = append(branches, line)
		}
	}
	return branches
}

func SetupPrsCommand() *cobra.Command {
	var outputJSON bool
	var useLocal bool

	cmd := &cobra.Command{
		Use:   "prs [branchA] [branchB]",
		Short: "List PRs in branchA that are not yet in branchB",
		Args:  cobra.MaximumNArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if !utils.IsInsideGitRepository() {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}

			allBranches := GetBranchNames()
			var matches []string
			for _, b := range allBranches {
				if strings.HasPrefix(b, toComplete) {
					matches = append(matches, b)
				}
			}
			return matches, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !utils.IsInsideGitRepository() {
				return fmt.Errorf("not inside a git repository")
			}

			branchA := "dev"
			branchB := "rtm"

			if len(args) > 0 {
				branchA = args[0]
			}
			if len(args) > 1 {
				branchB = args[1]
			}

			prList, err := ListPullRequestsBetweenBranches(branchA, branchB, useLocal)
			if err != nil {
				return err
			}

			if outputJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(prList)
			}

			if len(prList) == 0 {
				fmt.Printf("No PRs found in '%s' but not in '%s'\n", branchA, branchB)
				return nil
			}

			fmt.Printf("PRs in '%s' but not in '%s':\n", branchA, branchB)
			for _, pr := range prList {
				fmt.Printf("#%s: %s\n", pr.Number, pr.Title)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output results as JSON")
	cmd.Flags().BoolVar(&useLocal, "local", false, "Use local branches instead of origin/<branch>")

	return cmd
}
