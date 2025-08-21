package prs

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/astein-peddi/git-tooling/utils"
	graphql "github.com/cli/shurcooL-graphql"
	"github.com/spf13/cobra"
)

func SetupPrsCommand() *cobra.Command {
	var branchA, branchB string

	cmd := &cobra.Command{
		Use:   "prs <branchA> <branchB>",
		Short: "List PRs merged into branchA that are not in branchB",
		Args:  cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if !utils.IsInsideGitRepository() {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
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
			if !utils.IsInsideGitRepository() {
				return fmt.Errorf("not inside a git repository")
			}

			branchA = args[0]
			branchB = args[1]

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			owner, repo, err := utils.GetRepoOwnerAndName()
			if err != nil {
				return err
			}

			debug, _ := cmd.Flags().GetBool("debug")

			var _ context.Context = ctx
			prs, err := fetchMergedPRsShurcooL(owner, repo, branchA, debug)
			if err != nil {
				return err
			}

			fmt.Printf("PRs in '%s' not in '%s':\n", branchA, branchB)
			for _, pr := range prs {
				if pr.MergeCommit.Oid == "" {
					continue
				}

				if !commitInBranch(pr.MergeCommit.Oid, branchB) {
					fmt.Printf("#%d: %s (%s)\n", pr.Number, pr.Title, pr.URL)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolP("debug", "d", false, "Enable debug output")

	return cmd
}

func fetchMergedPRsShurcooL(owner, repo, baseBranch string, debug bool) ([]PR, error) {
	client, err := utils.GetGhGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	var query struct {
		Repository struct {
			PullRequests struct {
				Nodes []struct {
					Number      int
					Title       string
					URL         string
					MergeCommit struct {
						Oid string
					}
				}
			} `graphql:"pullRequests(baseRefName: $baseRef, states: MERGED, first: 100)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]any{
		"owner":   graphql.String(owner),
		"repo":    graphql.String(repo),
		"baseRef": graphql.String(baseBranch),
	}

	if debug {
		fmt.Println("=== GraphQL Query (shurcooL) ===")
		fmt.Printf("Owner: %s, Repo: %s, BaseRef: %s\n", owner, repo, baseBranch)
		fmt.Println("===============================\n")
	}

	err = client.Query("MergedPullRequests", &query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch merged PRs: %w", err)
	}

	var prs []PR
	for _, node := range query.Repository.PullRequests.Nodes {
		prs = append(prs, PR{
			Number:      node.Number,
			Title:       node.Title,
			URL:         node.URL,
			MergeCommit: node.MergeCommit,
		})
	}

	fmt.Printf("Found %d merged PRs\n", len(prs))
	return prs, nil
}

func commitInBranch(sha, branch string) bool {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", sha, branch)

	return cmd.Run() == nil
}