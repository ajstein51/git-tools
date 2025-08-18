package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

type PR struct {
	Number string `json:"number"`
	Commit string `json:"commit"`
}

func getPRs(branchA, branchB string) ([]PR, error) {
	cmd := exec.Command("git", "log", branchA, "^"+branchB, "--merges", "--pretty=format:%H %s")
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git log error: %v\n%s", err, stderr.String())
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	prRegex := regexp.MustCompile(`Merge pull request #(\d+)`)

	var prs []PR
	for _, line := range lines {
		if match := prRegex.FindStringSubmatch(line); match != nil {
			prs = append(prs, PR{
				Number: match[1],
				Commit: line,
			})
		}
	}

	return prs, nil
}

func branchExists(branch string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", branch)
	if err := cmd.Run(); err == nil {
		return true
	}
	cmd = exec.Command("git", "rev-parse", "--verify", "origin/"+branch)
	return cmd.Run() == nil
}

func getBranchNames() []string {
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

func main() {
	var outputJSON bool

	rootCmd := &cobra.Command{
		Use:   "peddi-tooling",
		Short: "Peddi Tooling CLI",
		Long:  "Tooling for Peddi Git Tasks",
	}

	prsCmd := &cobra.Command{
		Use:   "prs [branchA] [branchB]",
		Short: "List PRs in branchA that are not yet in branchB",
		Args:  cobra.MaximumNArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// Return branch names that match what user typed
			allBranches := getBranchNames()
			var matches []string
			for _, b := range allBranches {
				if strings.HasPrefix(b, toComplete) {
					matches = append(matches, b)
				}
			}
			return matches, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			branchA := "dev"
			branchB := "rtm"

			if len(args) > 0 {
				branchA = args[0]
			}
			if len(args) > 1 {
				branchB = args[1]
			}

			if !branchExists(branchA) {
				return fmt.Errorf("development branch '%s' does not exist locally or on origin", branchA)
			}
			if !branchExists(branchB) {
				return fmt.Errorf("RTM branch '%s' does not exist locally or on origin", branchB)
			}

			prs, err := getPRs(branchA, branchB)
			if err != nil {
				return err
			}

			if outputJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(prs)
			}

			if len(prs) == 0 {
				fmt.Printf("No PRs found in '%s' but not in '%s'\n", branchA, branchB)
				return nil
			}

			fmt.Printf("PRs in '%s' but not in '%s':\n", branchA, branchB)
			for _, pr := range prs {
				fmt.Printf("#%s: %s\n", pr.Number, pr.Commit)
			}
			return nil
		},
	}

	prsCmd.Flags().BoolVar(&outputJSON, "json", false, "Output results as JSON")

	rootCmd.AddCommand(prsCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
