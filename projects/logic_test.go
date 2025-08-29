package projects

import (
	"fmt"
	"testing"
	"time"

	"github.com/astein-peddi/git-tooling/models"
	"github.com/cli/shurcooL-graphql"
	"github.com/stretchr/testify/assert"
)

func TestGetFieldValue(t *testing.T) {
	testCases := []struct {
		name     string
		item     ProjectItem
		expected string
	}{
		{
			name: "Single Select Value",
			item: ProjectItem{
				FieldValueByName: struct {
					Typename          graphql.String `graphql:"__typename"`
					SingleSelectValue struct{ Name string } `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
					TextValue         struct{ Text string } `graphql:"... on ProjectV2ItemFieldTextValue"`
				}{
					Typename:          "ProjectV2ItemFieldSingleSelectValue",
					SingleSelectValue: struct{ Name string }{Name: "High"},
				},
			},
			expected: "High",
		},
		{
			name:     "No Value",
			item:     ProjectItem{},
			expected: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, getFieldValue(tc.item))
		})
	}
}

func TestGetLinkedPRs(t *testing.T) {
	pr101 := newTestPR(101, "Feat: New API", nil)
	pr102 := newTestPR(102, "Fix: Bug", nil)
	testCases := []struct {
		name     string
		item     ProjectItem
		expected []PullRequestFragment
	}{
		{
			name:     "Item is a PR, should return empty",
			item:     newTestItemWithPR(pr101),
			expected: []PullRequestFragment{},
		},
		{
			name:     "Issue with multiple linked PRs, should be sorted",
			item:     newTestItemWithIssue(1, "An issue", pr101, pr102),
			expected: []PullRequestFragment{pr102, pr101},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := getLinkedPRs(tc.item)
			assert.Len(t, actual, len(tc.expected))
			if len(actual) > 0 && len(tc.expected) > 0 {
				assert.Equal(t, tc.expected[0].Number, actual[0].Number)
			}
		})
	}
}

func TestProcessProjectItems(t *testing.T) {
	mergedTime := time.Now().String()
	prMerged := newTestItemWithPR(newTestPR(101, "Merged PR", &mergedTime))
	prOpen := newTestItemWithPR(newTestPR(102, "Open PR", nil, "test-user"))
	issueWithPR := newTestItemWithIssue(1, "Issue with PR", newTestPR(103, "Linked PR", nil))
	issueNoPR := newTestItemWithIssue(2, "Issue with no PR")
	draft := newTestItemWithDraft("A draft idea")
	allItems := []ProjectItem {
		prMerged.withCustomField("Done"),
		prOpen.withCustomField("In Progress"),
		issueWithPR.withCustomField("In Progress"),
		issueNoPR.withCustomField("Todo"),
		draft.withCustomField(""),
	}

	noPRFilter := func(item ProjectItem) bool {
		if item.Content.Typename == "PullRequest" {
			return false
		}
		if item.Content.Typename == "Issue" {
			return len(getLinkedPRs(item)) == 0
		}
		return item.Content.Typename == "DraftIssue"
	}

	t.Run("Group by custom field", func(t *testing.T) {
		processed := processProjectItems(allItems, nil, "Status")
		assert.Len(t, processed, 5)
	})
	t.Run("Filter: no-pr", func(t *testing.T) {
		processed := processProjectItems(allItems, noPRFilter, "")
		assert.Len(t, processed, 2)
	})
}

func TestFetchProjectData(t *testing.T) {
	t.Run("Successfully finds org project", func(t *testing.T) {
		mockResponse := &struct {
			Organization struct {
				ProjectV2 *struct {
					Title string
					Items struct {
						Nodes    []ProjectItem
						PageInfo models.PageInfo
					} `graphql:"items(first: 100, after: $after)"`
				} `graphql:"projectV2(number: $number)"`
			} `graphql:"organization(login: $owner)"`
		}{}
		mockResponse.Organization.ProjectV2 = &struct {
			Title string
			Items struct {
				Nodes    []ProjectItem
				PageInfo models.PageInfo
			} `graphql:"items(first: 100, after: $after)"`
		}{
			Title: "My Org Project",
			Items: struct {
				Nodes    []ProjectItem
				PageInfo models.PageInfo
			}{
				Nodes: []ProjectItem{newTestItemWithDraft("Test Draft")},
			},
		}

		mockClient := &mockGQLClient{mockResponse: mockResponse}
		items, title, err := fetchProjectData(mockClient, "my-org", "my-repo", 1, "")
		assert.NoError(t, err)
		assert.Equal(t, "My Org Project", title)
		assert.Len(t, items, 1)
	})

	t.Run("API returns an error on org query", func(t *testing.T) {
		mockClient := &mockGQLClient{mockErr: fmt.Errorf("permission denied")}
		_, _, err := fetchProjectData(mockClient, "my-org", "my-repo", 1, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
	})
}