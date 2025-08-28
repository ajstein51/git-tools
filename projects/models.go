package projects

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/cli/shurcooL-graphql"
)

type ItemFilter func(ProjectItem) bool

type Model struct {
	repoOwner    string
	repoName     string
	projectTitle string
	groupByField string
	items        []ProjectItem

	table       table.Model
	dividerRows map[int]bool
}

type PullRequestFragment struct {
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
}

type ProjectItem struct {
	ID               graphql.ID
	FieldValueByName struct {
		Typename          graphql.String `graphql:"__typename"`
		SingleSelectValue struct {
			Name string
		} `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
		TextValue struct {
			Text string
		} `graphql:"... on ProjectV2ItemFieldTextValue"`
	} `graphql:"fieldValueByName(name: $fieldName)"`
	Content ProjectItemContent `graphql:"content"`
}

type ProjectItemContent struct {
	Typename graphql.String `graphql:"__typename"`
	Issue    struct {
		Number        int
		Title         string
		TimelineItems struct {
			Nodes []struct {
				ConnectedEvent struct {
					Subject struct {
						PullRequest PullRequestFragment `graphql:"... on PullRequest"`
					}
				} `graphql:"... on ConnectedEvent"`
				CrossReferencedEvent struct {
					Source struct {
						PullRequest PullRequestFragment `graphql:"... on PullRequest"`
					}
				} `graphql:"... on CrossReferencedEvent"`
				ReferencedEvent struct {
					Subject struct {
						PullRequest PullRequestFragment `graphql:"... on PullRequest"`
					}
				} `graphql:"... on ReferencedEvent"`
			}
		} `graphql:"timelineItems(itemTypes: [CONNECTED_EVENT, CROSS_REFERENCED_EVENT, REFERENCED_EVENT], first: 5)"`
	} `graphql:"... on Issue"`
	PR         PullRequestFragment `graphql:"... on PullRequest"`
	DraftIssue struct {
		Title string
	} `graphql:"... on DraftIssue"`
}