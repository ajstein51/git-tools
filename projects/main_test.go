package projects

import (
	"fmt"
	"reflect"
)

type mockGQLClient struct {
	mockResponse any
	mockErr      error
}

func (m *mockGQLClient) Query(queryName string, response any, variables map[string]any) error {
	if m.mockErr != nil {
		return m.mockErr
	}
	if m.mockResponse == nil {
		return fmt.Errorf("mock response is nil for query %s", queryName)
	}
	val := reflect.ValueOf(response)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("response must be a pointer")
	}
	val.Elem().Set(reflect.ValueOf(m.mockResponse).Elem())
	return nil
}

func newTestPR(number int, title string, mergedAt *string, reviewerLogins ...string) PullRequestFragment {
	pr := PullRequestFragment{
		Number:   number,
		Title:    title,
		MergedAt: mergedAt,
	}
	for _, login := range reviewerLogins {
		pr.ReviewRequests.Nodes = append(pr.ReviewRequests.Nodes, struct {
			RequestedReviewer struct {
				OnUser struct {
					Login string
				} `graphql:"... on User"`
			} `graphql:"requestedReviewer"`
		}{
			RequestedReviewer: struct {
				OnUser struct {
					Login string
				} `graphql:"... on User"`
			}{
				OnUser: struct {
					Login string
				}{
					Login: login,
				},
			},
		})
	}
	return pr
}

func newTestItemWithPR(pr PullRequestFragment) ProjectItem {
	return ProjectItem{
		Content: ProjectItemContent{
			Typename: "PullRequest",
			PR:       pr,
		},
	}
}

func newTestItemWithIssue(number int, title string, linkedPRs ...PullRequestFragment) ProjectItem {
       item := ProjectItem{
	       Content: ProjectItemContent{
		       Typename: "Issue",
		       Issue: struct {
			       Number        int
			       Title         string
			       TimelineItems struct {
				       Nodes []struct {
					       ConnectedEvent       struct{ Subject struct{ PullRequest PullRequestFragment `graphql:"... on PullRequest"` } } `graphql:"... on ConnectedEvent"`
					       CrossReferencedEvent struct{ Source struct{ PullRequest PullRequestFragment `graphql:"... on PullRequest"` } }  `graphql:"... on CrossReferencedEvent"`
					       ReferencedEvent      struct{ Subject struct{ PullRequest PullRequestFragment `graphql:"... on PullRequest"` } } `graphql:"... on ReferencedEvent"`
				       }
			       } `graphql:"timelineItems(itemTypes: [CONNECTED_EVENT, CROSS_REFERENCED_EVENT, REFERENCED_EVENT], first: 5)"`
		       }{
			       Number: number,
			       Title:  title,
		       },
	       },
       }
       for _, pr := range linkedPRs {
	       item.Content.Issue.TimelineItems.Nodes = append(item.Content.Issue.TimelineItems.Nodes, struct {
		       ConnectedEvent       struct{ Subject struct{ PullRequest PullRequestFragment `graphql:"... on PullRequest"` } } `graphql:"... on ConnectedEvent"`
		       CrossReferencedEvent struct{ Source struct{ PullRequest PullRequestFragment `graphql:"... on PullRequest"` } }  `graphql:"... on CrossReferencedEvent"`
		       ReferencedEvent      struct{ Subject struct{ PullRequest PullRequestFragment `graphql:"... on PullRequest"` } } `graphql:"... on ReferencedEvent"`
	       }{
		       CrossReferencedEvent: struct{ Source struct{ PullRequest PullRequestFragment `graphql:"... on PullRequest"` } }{
			       Source: struct{ PullRequest PullRequestFragment `graphql:"... on PullRequest"` }{
				       PullRequest: pr,
			       },
		       },
	       })
       }
       return item
}

func newTestItemWithDraft(title string) ProjectItem {
       return ProjectItem{
	       Content: ProjectItemContent{
		       Typename: "DraftIssue",
		       DraftIssue: struct {
			       Title string
		       }{
			       Title: title,
		       },
	       },
       }
}

func (p ProjectItem) withCustomField(value string) ProjectItem {
	p.FieldValueByName.Typename = "ProjectV2ItemFieldSingleSelectValue"
	p.FieldValueByName.SingleSelectValue.Name = value
	return p
}