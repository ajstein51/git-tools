package projects

import (
	"fmt"

	"github.com/astein-peddi/git-tooling/utils"
	graphql "github.com/cli/shurcooL-graphql"
	"github.com/spf13/cobra"
)

func SetupProjectsCommand() *cobra.Command {
	var projectNumber int

	cmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage GitHub Projects",
		Long:  "Commands to list and filter issues/cards from GitHub Projects",
	}

	cmd.PersistentFlags().IntVar(&projectNumber, "id", 0, "Project number (defaults to last project)")

	listCmd := &cobra.Command{
		Use: "list",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if projectNumber == 0 {
				num, err := getLastProjectNumber()
				if err != nil {
					return fmt.Errorf("failed to get last project number: %w", err)
				}
				projectNumber = num
			}
			return nil
		},
	}

	listAllCmd := &cobra.Command{
		Use:   "all",
		Short: "List all issues/cards in the project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listProjectCards(projectNumber, nil)
		},
	}

	listNoPRCmd := &cobra.Command{
		Use:   "no-pr",
		Short: "List issues with no PR",
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := func(item ProjectItem) bool {
				return item.Content.Typename == "Issue" && item.Content.Issue.TimelineItems.TotalCount == 0
			}

			return listProjectCards(projectNumber, filter)
		},
	}

	listPRCmd := &cobra.Command{
		Use:   "with-pr",
		Short: "List issues with a PR",
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := func(item ProjectItem) bool {
				isPR := item.Content.Typename == "PullRequest"
				isIssueWithPR := item.Content.Typename == "Issue" && item.Content.Issue.TimelineItems.TotalCount > 0
				return isPR || isIssueWithPR
			}

			return listProjectCards(projectNumber, filter)
		},
	}

	listPRNotMergedCmd := &cobra.Command{
		Use:   "pr-not-merged",
		Short: "List issues with PR that is not merged",
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := func(item ProjectItem) bool {
				return item.Content.Typename == "Issue" && item.Content.PR.MergedAt == nil
			}

			return listProjectCards(projectNumber, filter)
		},
	}

	listReviewerCmd := &cobra.Command{
		Use:   "reviewer",
		Short: "List issues where you are a reviewer",
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := func(item ProjectItem) bool {
				if item.Content.Typename != "PullRequest" {
					return false
				}

				username, _ := utils.GetGhUsernameGraphQL()
				for _, rr := range item.Content.PR.ReviewRequests.Nodes {
					if rr.RequestedReviewer.OnUser.Login == username {
						return true
					}
				}
				return false
			}

			return listProjectCards(projectNumber, filter)
		},
	}

	listCmd.AddCommand(listAllCmd, listNoPRCmd, listPRCmd, listPRNotMergedCmd, listReviewerCmd)
	cmd.AddCommand(listCmd)

	return cmd
}

func getLastProjectNumber() (int, error) {
	client, err := utils.GetGhGraphQLClient()
	if err != nil {
		return 0, err
	}

	var query struct {
		Viewer struct {
			Projects struct {
				Nodes []struct {
					Number int
				}
			} `graphql:"projectsV2(first: 50, orderBy: {field: UPDATED_AT, direction: DESC})"`
		}
	}

	err = client.Query("LastProjectNumber", &query, nil)
	if err != nil {
		return 0, err
	}

	if len(query.Viewer.Projects.Nodes) == 0 {
		return 0, fmt.Errorf("no projects found")
	}

	return query.Viewer.Projects.Nodes[0].Number, nil
}

func listProjectCards(projectNumber int, filter func(ProjectItem) bool) error {
	client, err := utils.GetGhGraphQLClient()
	if err != nil {
		return err
	}

	var query struct {
		Viewer struct {
			ProjectV2 struct {
				Items struct {
					Nodes []ProjectItem
				} `graphql:"items(first: 100)"`
			} `graphql:"projectV2(number: $number)"`
		}
	}

	variables := map[string]any{
		"number": graphql.Int(projectNumber),
	}

	err = client.Query("ProjectItems", &query, variables)
	if err != nil {
		return err
	}

	for _, node := range query.Viewer.ProjectV2.Items.Nodes {
		if filter == nil || filter(node) {
			switch node.Content.Typename {
			case "Issue":
				fmt.Printf("Issue #%d: %s\n", node.Content.Issue.Number, node.Content.Issue.Title)
			case "PullRequest":
				fmt.Printf("PR #%d: %s\n", node.Content.PR.Number, node.Content.PR.Title)
			case "DraftIssue":
				fmt.Printf("Draft: %s\n", node.Content.DraftIssue.Title)
			}
		}
	}

	return nil
}

type ProjectItem struct {
	ID      graphql.ID
	Content struct {
		Typename graphql.String `graphql:"__typename"`

		Issue struct {
			Number int
			Title  string
			TimelineItems struct {
				TotalCount int
			} `graphql:"timelineItems(itemTypes: [CROSS_REFERENCED_EVENT, CONNECTED_EVENT], first: 1)"`
		} `graphql:"... on Issue"`

		PR struct {
			Number   int
			Title    string
			MergedAt *string
			ReviewRequests struct {
				Nodes []struct {
					RequestedReviewer struct {
						OnUser struct {
							Login string
						} `graphql:"... on User"`
					} `graphql:"requestedReviewer"`
				}
			} `graphql:"reviewRequests(first: 10)"`
		} `graphql:"... on PullRequest"`

		DraftIssue struct {
			Title string
		} `graphql:"... on DraftIssue"`
	} `graphql:"content"`
}
