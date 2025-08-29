package prs

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/astein-peddi/git-tooling/loader"
	"github.com/astein-peddi/git-tooling/utils"
	"github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func SetupPrsCommand() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "prs <sourceBranch> <targetBranch>",
		Short: "List PRs in a source branch that are not in a target branch",
		Args:  cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if !utils.IsInsideGitRepository() {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			allBranches := utils.GetBranchNames()
			var matches []string
			for _, b := range allBranches {
				if strings.HasPrefix(b, toComplete) {
					matches = append(matches, b)
				}
			}
			return matches, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			branchA := args[0]
			branchB := args[1]
			isLocal, _ := cmd.Flags().GetBool("local")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			if isLocal {
				return fmt.Errorf("local mode is not yet implemented")
			}
			
			if !utils.DoesBranchExist(branchA, isLocal) {
				return fmt.Errorf("branch '%s' does not exist", branchA)
			}
			if !utils.DoesBranchExist(branchB, isLocal) {
				return fmt.Errorf("branch '%s' does not exist", branchB)
			}
			
			owner, repo, err := utils.GetRepoOwnerAndName()
			if err != nil {
				return err
			}

			task := func() (any, error) {
				client, err := utils.GetGhGraphQLClient()
				if err != nil {
					return nil, err
				}

				var wg sync.WaitGroup
				resultsChan := make(chan branchScanResult, 2)
				
				for _, branch := range []string{branchA, branchB} {
					wg.Add(1)
					go func(b string) {
						defer wg.Done()
						prs, err := fetchPRsForBranch(client, owner, repo, b, limit)
						resultsChan <- branchScanResult{branchName: b, prs: prs, err: err}
					}(branch)
				}

				wg.Wait()
				close(resultsChan)
				
				var branchAPrs, branchBPrs []PR
				for result := range resultsChan {
					if result.err != nil {
						return nil, result.err
					}
					if result.branchName == branchA {
						branchAPrs = result.prs
					} else {
						branchBPrs = result.prs
					}
				}
				
				branchBSet := make(map[int]bool)
				for _, pr := range branchBPrs {
					branchBSet[pr.Number] = true
				}

				var finalResults []PR
				for _, prA := range branchAPrs {
					if !branchBSet[prA.Number] {
						finalResults = append(finalResults, prA)
					}
				}

				return finalResults, nil
			}

			result, err := loader.Run("Scanning branch histories", task)
			if err != nil {
				return err
			}
			
			finalPRs := result.([]PR)

			if jsonOutput {
				jsonData, err := json.MarshalIndent(finalPRs, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal results to JSON: %w", err)
				}
				fmt.Println(string(jsonData))
				return nil
			}
			
			p := tea.NewProgram(initialModel(branchA, branchB, finalPRs), tea.WithAltScreen())
			_, err = p.Run()
			
			return err
		},
	}

	cmd.Flags().BoolP("local", "l", false, "Compare local branches")
	cmd.Flags().IntVar(&limit, "limit", 0, "Max number of commits to scan per branch (0=all)")
	cmd.Flags().Bool("json", false, "Output results in JSON format")

	return cmd
}