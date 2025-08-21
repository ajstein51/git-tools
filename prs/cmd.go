package prs

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/astein-peddi/git-tooling/utils"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/shurcooL-graphql"
	"github.com/spf13/cobra"
)

func SetupPrsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prs <branchA> <branchB>",
		Short: "List PRs merged into branchA that are not in branchB",
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

			owner, repo, err := utils.GetRepoOwnerAndName()
			if err != nil {
				return err
			}

			client, err := utils.GetGhGraphQLClient()
			if err != nil {
				return err
			}

			var cursor *graphql.String 
			foundDiffOnAnyPage := false

			fmt.Printf("Comparing recent PRs merged into '%s' that are not yet in '%s'...\n\n", branchA, branchB)

			for {
				prs, pageInfo, err := fetchRecentMergedPRsPage(client, owner, repo, branchA, cursor)
				if err != nil {
					return err
				}

				if len(prs) == 0 && cursor == nil {
					fmt.Printf("No recently merged PRs found for branch '%s'.\n", branchA)
					break
				}

				foundDiffOnThisPage := false
				for _, pr := range prs {
					if pr.MergeCommit.Oid == "" {
						continue
					}
					if !commitInBranch(pr.MergeCommit.Oid, branchB) {
						fmt.Printf("#%d: %s (%s)\n", pr.Number, pr.Title, pr.URL)
						foundDiffOnThisPage = true
						foundDiffOnAnyPage = true
					}
				}

				if !foundDiffOnThisPage {
					fmt.Println("(No differences found on this page)")
				}

				if !pageInfo.HasNextPage {
					break
				}

				fmt.Print("\n--- Press Enter for next page, or q to quit: ")
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				if strings.TrimSpace(input) == "q" {
					break
				}

				newCursor := graphql.String(pageInfo.EndCursor)
				cursor = &newCursor
			}

			if !foundDiffOnAnyPage {
				fmt.Println("\nFinished. No differences found between the branches for recent PRs.")
			}

			return nil
		},
	}
	return cmd
}

func fetchRecentMergedPRsPage(client *api.GraphQLClient, owner, repo, baseBranch string, afterCursor *graphql.String) ([]PR, PageInfo, error) {
	var query struct {
		Repository struct {
			PullRequests struct {
				Nodes    []PR
				PageInfo PageInfo
			} `graphql:"pullRequests(baseRefName: $baseRef, states: MERGED, first: 20, after: $after, orderBy: {field: UPDATED_AT, direction: DESC})"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]any{
		"owner":   graphql.String(owner),
		"repo":    graphql.String(repo),
		"baseRef": graphql.String(baseBranch),
		"after":   afterCursor, // This is now the correct type
	}

	err := client.Query("MergedPullRequests", &query, variables)
	if err != nil {
		return nil, PageInfo{}, fmt.Errorf("failed to fetch merged PRs: %w", err)
	}

	return query.Repository.PullRequests.Nodes, query.Repository.PullRequests.PageInfo, nil
}

func commitInBranch(sha, branch string) bool {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", sha, branch)
	
	return cmd.Run() == nil
}