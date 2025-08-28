package projects

import (
	"fmt"
	"sort"

	"github.com/astein-peddi/git-tooling/models"
	"github.com/cli/shurcooL-graphql"
)

func fetchProjectData(client models.GQLClient, owner, repo string, projectNumber int, groupByField string) ([]ProjectItem, string, error) {
	var allItems []ProjectItem
	var projectTitle string

	var orgQuery struct {
		Organization struct {
			ProjectV2 *struct {
				Title string
				Items struct {
					Nodes    []ProjectItem
					PageInfo models.PageInfo
				} `graphql:"items(first: 100, after: $after)"`
			} `graphql:"projectV2(number: $number)"`
		} `graphql:"organization(login: $owner)"`
	}

	orgVariables := map[string]any{
		"owner":     graphql.String(owner),
		"number":    graphql.Int(projectNumber),
		"after":     (*graphql.String)(nil),
		"fieldName": graphql.String(groupByField),
	}

	orgErr := client.Query("OrgProjectItems", &orgQuery, orgVariables)
	if orgErr != nil {
		return nil, "", fmt.Errorf("error querying organization project: %w", orgErr)
	}

	if orgQuery.Organization.ProjectV2 != nil {
		projectTitle = orgQuery.Organization.ProjectV2.Title
		allItems = append(allItems, orgQuery.Organization.ProjectV2.Items.Nodes...)
		pageInfo := orgQuery.Organization.ProjectV2.Items.PageInfo

		for pageInfo.HasNextPage {
			orgVariables["after"] = graphql.String(pageInfo.EndCursor)

			if err := client.Query("OrgProjectItems", &orgQuery, orgVariables); err != nil {
				return nil, "", fmt.Errorf("failed during pagination of org project items: %w", err)
			}

			allItems = append(allItems, orgQuery.Organization.ProjectV2.Items.Nodes...)
			pageInfo = orgQuery.Organization.ProjectV2.Items.PageInfo
		}

		return allItems, projectTitle, nil
	}

	var repoQuery struct {
		Repository struct {
			ProjectV2 *struct {
				Title string
				Items struct {
					Nodes    []ProjectItem
					PageInfo models.PageInfo
				} `graphql:"items(first: 100, after: $after)"`
			} `graphql:"projectV2(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	repoVariables := map[string]any{
		"owner":     graphql.String(owner),
		"repo":      graphql.String(repo),
		"number":    graphql.Int(projectNumber),
		"after":     (*graphql.String)(nil),
		"fieldName": graphql.String(groupByField),
	}

	repoErr := client.Query("RepoProjectItems", &repoQuery, repoVariables)
	if repoErr != nil {
		return nil, "", fmt.Errorf("error querying repository project: %w", repoErr)
	}

	if repoQuery.Repository.ProjectV2 != nil {
		projectTitle = repoQuery.Repository.ProjectV2.Title
		allItems = append(allItems, repoQuery.Repository.ProjectV2.Items.Nodes...)
		pageInfo := repoQuery.Repository.ProjectV2.Items.PageInfo

		for pageInfo.HasNextPage {
			repoVariables["after"] = graphql.String(pageInfo.EndCursor)
			if err := client.Query("RepoProjectItems", &repoQuery, repoVariables); err != nil {
				return nil, "", fmt.Errorf("failed during pagination of repo project items: %w", err)
			}

			allItems = append(allItems, repoQuery.Repository.ProjectV2.Items.Nodes...)
			pageInfo = repoQuery.Repository.ProjectV2.Items.PageInfo
		}

		return allItems, projectTitle, nil
	}

	return nil, "", fmt.Errorf("failed to find project #%d. Please check the project ID and your permissions", projectNumber)
}

func processProjectItems(items []ProjectItem, filter itemFilter, groupByField string) []ProjectItem {
	var filteredItems []ProjectItem
	for _, item := range items {
		if filter == nil || filter(item) {
			filteredItems = append(filteredItems, item)
		}
	}

	if groupByField != "" {
		sort.SliceStable(filteredItems, func(i, j int) bool {
			fieldValueI := getFieldValue(filteredItems[i])
			fieldValueJ := getFieldValue(filteredItems[j])
			if fieldValueI == "" && fieldValueJ != "" {
				return false
			}
			if fieldValueI != "" && fieldValueJ == "" {
				return true
			}

			return fieldValueI < fieldValueJ
		})
	} else {
		sort.Slice(filteredItems, func(i, j int) bool {
			return i < j
		})
	}

	return filteredItems
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

func getFieldValue(item ProjectItem) string {
	if item.FieldValueByName.Typename == "ProjectV2ItemFieldSingleSelectValue" {
		return item.FieldValueByName.SingleSelectValue.Name
	}
	if item.FieldValueByName.Typename == "ProjectV2ItemFieldTextValue" {
		return item.FieldValueByName.TextValue.Text
	}
	
	return ""
}