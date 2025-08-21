package projects

import (
	"fmt"
	"sort"

	"github.com/astein-peddi/git-tooling/models"
	"github.com/astein-peddi/git-tooling/utils"
	"github.com/cli/shurcooL-graphql"
	"github.com/spf13/cobra"
)

func SetupProjectsCommand() *cobra.Command {
	var projectNumber int
	var repoOwner, repoName string

	cmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage GitHub Projects",
		Long:  "Commands to list and filter issues/cards from GitHub Projects",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			repoOwner, repoName, err = utils.GetRepoOwnerAndName()
			if err != nil {
				return fmt.Errorf("failed to get repository details: %w", err)
			}

			if projectNumber == 0 {
				num, err := getLastProjectNumber(repoOwner, repoName)
				if err != nil {
					return fmt.Errorf("failed to get last project number: %w", err)
				}

				projectNumber = num
			}

			return nil
		},
	}

	cmd.PersistentFlags().IntVar(&projectNumber, "id", 0, "Project number (defaults to last project)")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Commands to list items from a project",
	}

	listAllCmd := &cobra.Command{
		Use:   "all",
		Short: "List all issues/cards in the project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listProjectCards(repoOwner, repoName, projectNumber, nil)
		},
	}

	listNoPRCmd := &cobra.Command{
		Use:   "no-pr",
		Short: "List items with no associated PR",
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := func(item ProjectItem) bool {
				if item.Content.Typename == "PullRequest" {
					return false // It is a PR.
				}
				if item.Content.Typename == "Issue" {
					return len(getLinkedPRs(item)) == 0 
				}

				return item.Content.Typename == "DraftIssue" 
			}

			return listProjectCards(repoOwner, repoName, projectNumber, filter)
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

			return listProjectCards(repoOwner, repoName, projectNumber, filter)
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

			return listProjectCards(repoOwner, repoName, projectNumber, filter)
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

			return listProjectCards(repoOwner, repoName, projectNumber, filter)
		},
	}

	listReviewerCmd.Flags().StringP("name", "n", "", "Filter by a specific GitHub username (defaults to the authenticated user)")

	listCmd.AddCommand(listAllCmd, listNoPRCmd, listPRCmd, listPRNotMergedCmd, listReviewerCmd)
	cmd.AddCommand(listCmd)

	return cmd
}

func getLastProjectNumber(owner, repo string) (int, error) {
	client, err := utils.GetGhGraphQLClient()
	if err != nil {
		return 0, err
	}

	var query struct {
		Repository struct {
			ProjectsV2 struct {
				Nodes []struct{ Number int }
			} `graphql:"projectsV2(first: 1, orderBy: {field: UPDATED_AT, direction: DESC})"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]any{"owner": graphql.String(owner), "repo": graphql.String(repo)}
	err = client.Query("RepoLastProjectNumber", &query, variables)
	if err != nil {
		return 0, err
	}
	if len(query.Repository.ProjectsV2.Nodes) == 0 {
		return 0, fmt.Errorf("no projects found in repository %s/%s", owner, repo)
	}

	return query.Repository.ProjectsV2.Nodes[0].Number, nil
}

func listProjectCards(owner, repo string, projectNumber int, filter func(ProjectItem) bool) error {
	client, err := utils.GetGhGraphQLClient()
	if err != nil {
		return err
	}

	var allItems []ProjectItem
	var foundProject bool

	var orgQuery struct {
		Organization struct {
			ProjectV2 *struct { 
				Items struct {
					Nodes    []ProjectItem
					PageInfo models.PageInfo
				} `graphql:"items(first: 100, after: $after)"`
			} `graphql:"projectV2(number: $number)"`
		} `graphql:"organization(login: $owner)"`
	}

	orgVariables := map[string]any{
		"owner":  graphql.String(owner),
		"number": graphql.Int(projectNumber),
		"after":  (*graphql.String)(nil),
	}

	if err := client.Query("OrgProjectItems", &orgQuery, orgVariables); err == nil && orgQuery.Organization.ProjectV2 != nil {
		foundProject = true

		allItems = append(allItems, orgQuery.Organization.ProjectV2.Items.Nodes...)
		pageInfo := orgQuery.Organization.ProjectV2.Items.PageInfo
		for pageInfo.HasNextPage {
			orgVariables["after"] = graphql.String(pageInfo.EndCursor)

			if err := client.Query("OrgProjectItems", &orgQuery, orgVariables); err != nil {
				break
			}

			allItems = append(allItems, orgQuery.Organization.ProjectV2.Items.Nodes...)
			pageInfo = orgQuery.Organization.ProjectV2.Items.PageInfo
		}
	}

	if !foundProject {
		var repoQuery struct {
			Repository struct {
				ProjectV2 *struct {
					Items struct {
						Nodes    []ProjectItem
						PageInfo models.PageInfo
					} `graphql:"items(first: 100, after: $after)"`
				} `graphql:"projectV2(number: $number)"`
			} `graphql:"repository(owner: $owner, name: $repo)"`
		}

		repoVariables := map[string]any{
			"owner":  graphql.String(owner),
			"repo":   graphql.String(repo),
			"number": graphql.Int(projectNumber),
			"after":  (*graphql.String)(nil),
		}
		
		if err := client.Query("RepoProjectItems", &repoQuery, repoVariables); err == nil && repoQuery.Repository.ProjectV2 != nil {
			foundProject = true

			allItems = append(allItems, repoQuery.Repository.ProjectV2.Items.Nodes...)
			pageInfo := repoQuery.Repository.ProjectV2.Items.PageInfo
			for pageInfo.HasNextPage {
				repoVariables["after"] = graphql.String(pageInfo.EndCursor)
				if err := client.Query("RepoProjectItems", &repoQuery, repoVariables); err != nil {
					break
				}

				allItems = append(allItems, repoQuery.Repository.ProjectV2.Items.Nodes...)
				pageInfo = repoQuery.Repository.ProjectV2.Items.PageInfo
			}
		}
	}

	if !foundProject {
		return fmt.Errorf("failed to find project #%d in organization '%s' or repository '%s/%s'. Please check the project ID and your permissions", projectNumber, owner, owner, repo)
	}

	var filteredItems []ProjectItem
	for _, item := range allItems {
		if filter == nil || filter(item) {
			filteredItems = append(filteredItems, item)
		}
	}

	sort.Slice(filteredItems, func(i, j int) bool {
		itemI := filteredItems[i]
		itemJ := filteredItems[j]
		isPRI := itemI.Content.Typename == "PullRequest"
		isPRJ := itemJ.Content.Typename == "PullRequest"
		
		if isPRI && isPRJ {
			return itemI.Content.PR.Number > itemJ.Content.PR.Number
		}
		if isPRI && !isPRJ {
			return true
		}
		if !isPRI && isPRJ {
			return false
		}

		return false
	})

	for _, node := range filteredItems {
		switch node.Content.Typename {
		case "Issue":
			fmt.Printf("Issue #%d: %s\n", node.Content.Issue.Number, node.Content.Issue.Title)
		case "PullRequest":
			fmt.Printf("PR #%d: %s\n", node.Content.PR.Number, node.Content.PR.Title)
		case "DraftIssue":
			fmt.Printf("Draft: %s\n", node.Content.DraftIssue.Title)
		}
	}

	return nil
}

func getLinkedPRs(item ProjectItem) []PullRequestFragment {
	var prs []PullRequestFragment
	if item.Content.Typename == "Issue" {
		for _, event := range item.Content.Issue.TimelineItems.Nodes {
			if event.ConnectedEvent.Subject.PullRequest.Number != 0 {
				prs = append(prs, event.ConnectedEvent.Subject.PullRequest)
			} else if event.CrossReferencedEvent.Source.PullRequest.Number != 0 {
				prs = append(prs, event.CrossReferencedEvent.Source.PullRequest)
			} else if event.ReferencedEvent.Subject.PullRequest.Number != 0 {
				prs = append(prs, event.ReferencedEvent.Subject.PullRequest)
			}
		}
	}
	
	sort.Slice(prs, func(i, j int) bool {
		return prs[i].Number > prs[j].Number
	})

	return prs
}
