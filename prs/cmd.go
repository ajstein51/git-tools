package prs

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/astein-peddi/git-tooling/models"
	"github.com/astein-peddi/git-tooling/utils"
	"github.com/cli/shurcooL-graphql"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var pullRequestRegex = regexp.MustCompile(`\(#(\d+)\)`)

func SetupPrsCommand() *cobra.Command {
	var limit int
	var pageSize int

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
			isLocal := cmd.Flags().Lookup("local").Value.String() == "true"

			if !utils.DoesBranchExist(branchA, isLocal) {
				return fmt.Errorf("branch '%s' does not exist or is not accessible", branchA)
			}
			if !utils.DoesBranchExist(branchB, isLocal) {
				return fmt.Errorf("branch '%s' does not exist or is not accessible", branchB)
			}

			owner, repo, err := utils.GetRepoOwnerAndName()
			if err != nil {
				return err
			}

			fmt.Printf("Comparing PRs merged into '%s' that are not yet in '%s'...\n\n", branchA, branchB)

			if isLocal {
				return fmt.Errorf("local mode not implemented")
			}

			client, err := utils.GetGhGraphQLClient()
			if err != nil {
				return err
			}

			branchAPrs, err := fetchPRsForBranch(client, owner, repo, branchA, limit)
			if err != nil {
				return err
			}

			branchBPrs, err := fetchPRsForBranch(client, owner, repo, branchB, limit)
			if err != nil {
				return err
			}

			branchBSet := make(map[int]bool)
			for _, pr := range branchBPrs {
				branchBSet[pr.Number] = true
			}

			var results []PR
			for _, prA := range branchAPrs {
				if !branchBSet[prA.Number] {
					results = append(results, prA)
				}
			}

			showPaginatedResultsInteractive(results, pageSize)

			return nil
		},
	}

	cmd.Flags().BoolP("local", "l", false, "Compare local branches")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of commits to scan per branch (0 = no limit)")
	cmd.Flags().IntVar(&pageSize, "page-size", 50, "Number of results to show per page")

	return cmd
}

func showPaginatedResultsInteractive(results []PR, pageSize int) {
	if pageSize <= 0 {
		pageSize = 50
	}

	total := len(results)

	if total < pageSize {
		for _, pr := range results {
			fmt.Printf("PR #%d '%s'\n", pr.Number, pr.Title)
		}

		return
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("Failed to enter raw mode:", err)
		return
	}

	defer term.Restore(int(os.Stdin.Fd()), oldState)

	start := 0
	for {
		end := min(start+pageSize, total)

		for _, pr := range results[start:end] {
			fmt.Printf("PR #%d '%s'\n", pr.Number, pr.Title)
		}

		fmt.Printf("\n-- Showing %d-%d of %d. Use 'h' to navigate left, 'l' to navigate right, 'q' to quit --\n", start+1, end, total)

		var b = make([]byte, 1)
		_, err := os.Stdin.Read(b)
		if err != nil {
			break
		}

		switch b[0] {
			case 'q':
				return
			case 'l':
				if end < total {
					start += pageSize
					fmt.Println()
				}
			case 'h':
				if start > 0 {
					start -= pageSize
					if start < 0 {
						start = 0
					}

					fmt.Println()
				}
			default:
		}

		fmt.Print("\033[H\033[2J") // clear screen, maybe change to a full terminal lib?
	}
}

func fetchCommitsInBranch(client models.GQLClient, owner, repo, branch string, limit int) ([]Commit, error) {
	var commits []Commit
	var cursor *string
	count := 0

	fmt.Fprintf(os.Stderr, "Scanning commits on '%s': ", branch)

	for {
		var query struct {
			Repository struct {
				Ref struct {
					Target struct {
						Commit struct {
							History struct {
								Edges []struct {
									Node struct {
										Oid     string
										Message string
									}
								}
								PageInfo models.PageInfo
							} `graphql:"history(first: 100, after: $after)"`
						} `graphql:"... on Commit"`
					}
				} `graphql:"ref(qualifiedName: $branch)"`
			} `graphql:"repository(owner: $owner, name: $repo)"`
		}

		variables := map[string]any{
			"owner":  graphql.String(owner),
			"repo":   graphql.String(repo),
			"branch": graphql.String(branch),
			"after":  (*graphql.String)(cursor),
		}

		err := client.Query("CommitsInBranch", &query, variables)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch commits in branch: %w", err)
		}

		fmt.Fprint(os.Stderr, ".")

		if len(query.Repository.Ref.Target.Commit.History.Edges) == 0 {
			break
		}

		for _, edge := range query.Repository.Ref.Target.Commit.History.Edges {
			commits = append(commits, Commit{
				Oid:     edge.Node.Oid,
				Message: edge.Node.Message,
			})
			count++

			if limit > 0 && count >= limit {
				fmt.Fprintln(os.Stderr)

				return commits, nil
			}
		}

		if !query.Repository.Ref.Target.Commit.History.PageInfo.HasNextPage {
			break
		}

		cursor = &query.Repository.Ref.Target.Commit.History.PageInfo.EndCursor
	}

	fmt.Fprintln(os.Stderr)

	return commits, nil
}

func fetchPRsForBranch(client models.GQLClient, owner, repo, branch string, limit int) ([]PR, error) {
	commits, err := fetchCommitsInBranch(client, owner, repo, branch, limit)
	if err != nil {
		return nil, err
	}

	seen := make(map[int]bool)
	var prs []PR

	for _, commit := range commits {
		if prNum, ok := extractPRNumber(commit.Message); ok {
			if !seen[prNum] {
				// commit.Message will also contain the description separated by newlines
				line := strings.SplitN(commit.Message, "\n", 2)[0]

				prs = append(prs, PR{Number: prNum, Title: line})

				seen[prNum] = true
			}
		}
	}

	return prs, nil
}

func extractPRNumber(message string) (int, bool) {
	matches := pullRequestRegex.FindStringSubmatch(message)

	if len(matches) == 2 {
		var num int
		_, err := fmt.Sscanf(matches[1], "%d", &num)
		if err == nil {
			return num, true
		}
	}

	return 0, false
}
