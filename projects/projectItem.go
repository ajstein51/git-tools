package projects

import (
	"github.com/cli/shurcooL-graphql"
)

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
	Content struct {
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
	} `graphql:"content"`
}
