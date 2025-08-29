package projects

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/astein-peddi/git-tooling/loader"
	"github.com/astein-peddi/git-tooling/models"
	"github.com/astein-peddi/git-tooling/utils"
	"github.com/charmbracelet/bubbletea"
	"github.com/cli/shurcooL-graphql"
	"github.com/spf13/cobra"
)

type projectDataResult struct {
	items []ProjectItem
	title string
}

func SetupProjectsCommand() *cobra.Command {
	var projectNumber int
	var repoOwner, repoName, groupByField string

	cmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage GitHub Projects",
		Long:  "Commands to list and filter issues/cards from GitHub Projects",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			if cmd.Flags().Changed("help") {
				return nil
			}
			if !utils.IsInsideGitRepository() {
				return fmt.Errorf("this command must be run from inside a Git repository")
			}
			repoOwner, repoName, err = utils.GetRepoOwnerAndName()
			if err != nil {
				return fmt.Errorf("failed to get repository details: %w", err)
			}
			if projectNumber == 0 {
				task := func() (any, error) {
					client, err := utils.GetGhGraphQLClient()
					if err != nil {
						return 0, err
					}
					return getLastProjectNumber(client, repoOwner, repoName)
				}

				result, err := loader.Run("Fetching latest project ID", task)
				if err != nil {
					return fmt.Errorf("failed to get last project number: %w", err)
				}

				projectNumber = result.(int)
			}

			return nil
		},
	}

	cmd.PersistentFlags().IntVar(&projectNumber, "id", 0, "Project number (defaults to last project)")
	cmd.PersistentFlags().StringVar(&groupByField, "groupBy", "", "Group by a custom field (e.g., 'Priority')")
	cmd.PersistentFlags().Bool("json", false, "Output results in JSON format")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Commands to list items from a project",
	}

		runListCommand := func(cmd *cobra.Command, filter ItemFilter) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")

		task := func() (any, error) {
			client, err := utils.GetGhGraphQLClient()
			if err != nil {
				return nil, err
			}

			allItems, projectTitle, err := fetchProjectData(client, repoOwner, repoName, projectNumber, groupByField)
			if err != nil {
				return nil, err
			}

			processedItems := processProjectItems(allItems, filter, groupByField)
			return projectDataResult{items: processedItems, title: projectTitle}, nil
		}

		result, err := loader.Run("Fetching project items", task)
		if err != nil {
			return err
		}

		data := result.(projectDataResult)

		if jsonOutput {
			jsonData, err := json.MarshalIndent(data.items, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal results to JSON: %w", err)
			}

			fmt.Println(string(jsonData))

			return nil
		}

		p := tea.NewProgram(initialModel(repoOwner, repoName, data.title, data.items, groupByField), tea.WithAltScreen())
		_, err = p.Run()

		return err
	}

	listAllCmd := &cobra.Command{
		Use:   "all",
		Short: "List all issues/cards in the project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListCommand(cmd, nil)
		},
	}

	listNoPRCmd := &cobra.Command{
		Use:   "no-pr",
		Short: "List items with no associated PR",
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := func(item ProjectItem) bool {
				if item.Content.Typename == "PullRequest" {
					return false
				}
				if item.Content.Typename == "Issue" {
					return len(getLinkedPRs(item)) == 0
				}
				
				return item.Content.Typename == "DraftIssue"
			}

			return runListCommand(cmd, filter)
		},
	}

	listPRCmd := &cobra.Command{
		Use:   "with-pr",
		Short: "List items that have an associated PR",
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := func(item ProjectItem) bool {
				if item.Content.Typename == "PullRequest" {
					return true
				}
				if item.Content.Typename == "Issue" {
					return len(getLinkedPRs(item)) > 0
				}
				return false
			}
			return runListCommand(cmd, filter)
		},
	}

	listPRNotMergedCmd := &cobra.Command{
		Use:   "pr-not-merged",
		Short: "List items with an unmerged PR",
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := func(item ProjectItem) bool {
				if item.Content.Typename == "PullRequest" {
					return item.Content.PR.MergedAt == nil
				}
				if item.Content.Typename == "Issue" {
					for _, pr := range getLinkedPRs(item) {
						if pr.MergedAt == nil {
							return true
						}
					}
				}

				return false
			}

			return runListCommand(cmd, filter)
		},
	}

	listReviewerCmd := &cobra.Command{
		Use:   "reviewer",
		Short: "List items where you or a specified user is a reviewer",
		RunE: func(cmd *cobra.Command, args []string) error {
			reviewerName, _ := cmd.Flags().GetString("name")
			if reviewerName == "" {
				var err error
				reviewerName, err = utils.GetGhUsernameGraphQL()
				if err != nil {
					return fmt.Errorf("could not determine current user: %w", err)
				}
			}

			filter := func(item ProjectItem) bool {
				var prsToCheck []PullRequestFragment
				switch item.Content.Typename {
				case "PullRequest":
					prsToCheck = append(prsToCheck, item.Content.PR)
				case "Issue":
					prsToCheck = getLinkedPRs(item)
				}
				for _, pr := range prsToCheck {
					for _, rr := range pr.ReviewRequests.Nodes {
						if rr.RequestedReviewer.OnUser.Login == reviewerName {
							return true
						}
					}
				}

				return false
			}

			return runListCommand(cmd, filter)
		},
	}

	listReviewerCmd.Flags().StringP("name", "n", "", "Filter by a specific GitHub username (defaults to the authenticated user)")

	listCmd.AddCommand(listAllCmd, listNoPRCmd, listPRCmd, listPRNotMergedCmd, listReviewerCmd)
	cmd.AddCommand(listCmd)

	return cmd
}

func getLastProjectNumber(client models.GQLClient, owner, repo string) (int, error) {
	var query struct {
		Organization struct {
			ProjectsV2 struct {
				Nodes []struct {
					Number    int
					CreatedAt time.Time
				}
			} `graphql:"projectsV2(first: 1, orderBy: {field: CREATED_AT, direction: DESC})"`
		} `graphql:"organization(login: $owner)"`
		Repository struct {
			ProjectsV2 struct {
				Nodes []struct {
					Number    int
					CreatedAt time.Time
				}
			} `graphql:"projectsV2(first: 1, orderBy: {field: CREATED_AT, direction: DESC})"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]any{
		"owner": graphql.String(owner),
		"repo":  graphql.String(repo),
	}

	if err := client.Query("LastProjectNumber", &query, variables); err != nil {
		return 0, err
	}

	orgProjectExists := len(query.Organization.ProjectsV2.Nodes) > 0
	repoProjectExists := len(query.Repository.ProjectsV2.Nodes) > 0
	if !orgProjectExists && !repoProjectExists {
		return 0, fmt.Errorf("no projects found in organization or repository for '%s/%s'", owner, repo)
	}
	if orgProjectExists && !repoProjectExists {
		return query.Organization.ProjectsV2.Nodes[0].Number, nil
	}
	if !orgProjectExists && repoProjectExists {
		return query.Repository.ProjectsV2.Nodes[0].Number, nil
	}

	orgProject := query.Organization.ProjectsV2.Nodes[0]
	repoProject := query.Repository.ProjectsV2.Nodes[0]
	if orgProject.CreatedAt.After(repoProject.CreatedAt) {
		return orgProject.Number, nil
	}
	
	return repoProject.Number, nil
}