package projects

// This was straight vibe coded. Tests may or may not be comprehensive.

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/astein-peddi/git-tooling/models"
	"github.com/cli/shurcooL-graphql"
	"github.com/stretchr/testify/assert"
)

type mockGQLClient struct {
	mockResponse any
	mockErr      error
}

func (m *mockGQLClient) Query(queryName string, response any, variables map[string]any) error {
	if m.mockErr != nil {
		return m.mockErr
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
		Content: struct {
			Typename   graphql.String `graphql:"__typename"`
			Issue      struct {
				Number        int
				Title         string
				TimelineItems struct {
					Nodes []struct {
						ConnectedEvent       struct {
							Subject struct {
								PullRequest PullRequestFragment `graphql:"... on PullRequest"`
							}
						} `graphql:"... on ConnectedEvent"`
						CrossReferencedEvent struct {
							Source struct {
								PullRequest PullRequestFragment `graphql:"... on PullRequest"`
							}
						} `graphql:"... on CrossReferencedEvent"`
						ReferencedEvent      struct {
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
		}{
			Typename: "PullRequest",
			PR:       pr,
		},
	}
}

func newTestItemWithIssue(number int, title string, linkedPRs ...PullRequestFragment) ProjectItem {
	item := ProjectItem{
		Content: struct {
			Typename   graphql.String `graphql:"__typename"`
			Issue      struct {
				Number        int
				Title         string
				TimelineItems struct {
					Nodes []struct {
						ConnectedEvent       struct {
							Subject struct {
								PullRequest PullRequestFragment `graphql:"... on PullRequest"`
							}
						} `graphql:"... on ConnectedEvent"`
						CrossReferencedEvent struct {
							Source struct {
								PullRequest PullRequestFragment `graphql:"... on PullRequest"`
							}
						} `graphql:"... on CrossReferencedEvent"`
						ReferencedEvent      struct {
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
		}{
			Typename: "Issue",
			Issue: struct {
				Number        int
				Title         string
				TimelineItems struct {
					Nodes []struct {
						ConnectedEvent       struct {
							Subject struct {
								PullRequest PullRequestFragment `graphql:"... on PullRequest"`
							}
						} `graphql:"... on ConnectedEvent"`
						CrossReferencedEvent struct {
							Source struct {
								PullRequest PullRequestFragment `graphql:"... on PullRequest"`
							}
						} `graphql:"... on CrossReferencedEvent"`
						ReferencedEvent      struct {
							Subject struct {
								PullRequest PullRequestFragment `graphql:"... on PullRequest"`
							}
						} `graphql:"... on ReferencedEvent"`
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
			ConnectedEvent       struct {
				Subject struct {
					PullRequest PullRequestFragment `graphql:"... on PullRequest"`
				}
			} `graphql:"... on ConnectedEvent"`
			CrossReferencedEvent struct {
				Source struct {
					PullRequest PullRequestFragment `graphql:"... on PullRequest"`
				}
			} `graphql:"... on CrossReferencedEvent"`
			ReferencedEvent      struct {
				Subject struct {
					PullRequest PullRequestFragment `graphql:"... on PullRequest"`
				}
			} `graphql:"... on ReferencedEvent"`
		}{
			ConnectedEvent: struct {
				Subject struct {
					PullRequest PullRequestFragment `graphql:"... on PullRequest"`
				}
			}{
				Subject: struct {
					PullRequest PullRequestFragment `graphql:"... on PullRequest"`
				}{
					PullRequest: pr,
				},
			},
		})
	}
	return item
}

func newTestItemWithDraft(title string) ProjectItem {
	return ProjectItem{
		Content: struct {
			Typename   graphql.String `graphql:"__typename"`
			Issue      struct {
				Number        int
				Title         string
				TimelineItems struct {
					Nodes []struct {
						ConnectedEvent       struct {
							Subject struct {
								PullRequest PullRequestFragment `graphql:"... on PullRequest"`
							}
						} `graphql:"... on ConnectedEvent"`
						CrossReferencedEvent struct {
							Source struct {
								PullRequest PullRequestFragment `graphql:"... on PullRequest"`
							}
						} `graphql:"... on CrossReferencedEvent"`
						ReferencedEvent      struct {
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
		}{
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

// --- UNIT TESTS ---

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
					SingleSelectValue struct {
						Name string
					} `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
					TextValue         struct {
						Text string
					} `graphql:"... on ProjectV2ItemFieldTextValue"`
				}{
					Typename: "ProjectV2ItemFieldSingleSelectValue",
					SingleSelectValue: struct {
						Name string
					}{Name: "High"},
				},
			},
			expected: "High",
		},
		{
			name: "Text Value",
			item: ProjectItem{
				FieldValueByName: struct {
					Typename          graphql.String `graphql:"__typename"`
					SingleSelectValue struct {
						Name string
					} `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
					TextValue         struct {
						Text string
					} `graphql:"... on ProjectV2ItemFieldTextValue"`
				}{
					Typename: "ProjectV2ItemFieldTextValue",
					TextValue: struct {
						Text string
					}{Text: "Some text"},
				},
			},
			expected: "Some text",
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
			name:     "Issue with no linked PRs",
			item:     newTestItemWithIssue(1, "An issue"),
			expected: []PullRequestFragment{},
		},
		{
			name:     "Issue with one linked PR",
			item:     newTestItemWithIssue(1, "An issue", pr101),
			expected: []PullRequestFragment{pr101},
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
			for i := range actual {
				assert.Equal(t, tc.expected[i].Number, actual[i].Number)
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

	allItems := []ProjectItem{
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

	withPRFilter := func(item ProjectItem) bool {
		if item.Content.Typename == "PullRequest" {
			return true
		}
		if item.Content.Typename == "Issue" {
			return len(getLinkedPRs(item)) > 0
		}
		return false
	}

	testCases := []struct {
		name           string
		items          []ProjectItem
		filter         func(ProjectItem) bool
		groupByField   string
		expectedTitles []string 
	}{
		{
			name:           "No filter, no group by (sorted by type and number)",
			items:          allItems,
			filter:         nil,
			groupByField:   "",
			expectedTitles: []string{"Open PR", "Merged PR", "Issue with PR", "Issue with no PR", "A draft idea"},
		},
		{
			name:           "Filter: no-pr",
			items:          allItems,
			filter:         noPRFilter,
			groupByField:   "",
			expectedTitles: []string{"Issue with no PR", "A draft idea"},
		},
		{
			name:           "Filter: with-pr",
			items:          allItems,
			filter:         withPRFilter,
			groupByField:   "",
			expectedTitles: []string{"Open PR", "Merged PR", "Issue with PR"},
		},
		{
			name:           "Group by custom field",
			items:          allItems,
			filter:         nil,
			groupByField:   "Status",
			expectedTitles: []string{"Merged PR", "Open PR", "Issue with PR", "Issue with no PR", "A draft idea"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			processed := processProjectItems(tc.items, tc.filter, tc.groupByField)
			assert.Len(t, processed, len(tc.expectedTitles))

			var actualTitles []string
			for _, item := range processed {
				switch item.Content.Typename {
				case "PullRequest":
					actualTitles = append(actualTitles, item.Content.PR.Title)
				case "Issue":
					actualTitles = append(actualTitles, item.Content.Issue.Title)
				case "DraftIssue":
					actualTitles = append(actualTitles, item.Content.DraftIssue.Title)
				}
			}
			assert.Equal(t, tc.expectedTitles, actualTitles)
		})
	}
}

func TestDisplayProjectItems(t *testing.T) {
	items := []ProjectItem{
		newTestItemWithPR(newTestPR(102, "Open PR", nil)).withCustomField("In Progress"),
		newTestItemWithIssue(1, "My Issue").withCustomField("Todo"),
		newTestItemWithDraft("A draft").withCustomField("Todo"),
	}

	testCases := []struct {
		name          string
		groupByField  string
		expectedLines []string
	}{
		{
			name:         "No grouping",
			groupByField: "",
			expectedLines: []string{
				"Project #42 - Test Project",
				"--------------------------------------------------",
				"",
				"PR #102: Open PR",
				"Issue #1: My Issue",
				"Draft: A draft",
			},
		},
		{
			name:         "With grouping",
			groupByField: "Status",
			expectedLines: []string{
				"Project #42 - Test Project",
				"--------------------------------------------------",
				"",
				"[In Progress] PR #102: Open PR",
				"",
				"[Todo] Issue #1: My Issue",
				"[Todo] Draft: A draft",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			displayProjectItems(items, 42, "Test Project", tc.groupByField)

			w.Close()
			os.Stdout = old
			var buf bytes.Buffer
			io.Copy(&buf, r)

			output := strings.TrimSpace(buf.String())
			expectedOutput := strings.Join(tc.expectedLines, "\n")

			normalizedOutput := strings.ReplaceAll(output, "\r\n", "\n")

			assert.Equal(t, expectedOutput, normalizedOutput)
		})
	}
}

func TestGetLastProjectNumber(t *testing.T) {
	t.Run("Org project is newer", func(t *testing.T) {
		mockResponse := &struct {
			Organization struct {
				ProjectsV2 struct {
					Nodes []struct {
						Number    int
						CreatedAt time.Time
					}
				} `graphql:"projectsV2(first: 1, orderBy: {field: CREATED_AT, direction: DESC})"`
			} `graphql:"organization(login: $owner)"`
			Repository   struct {
				ProjectsV2 struct {
					Nodes []struct {
						Number    int
						CreatedAt time.Time
					}
				} `graphql:"projectsV2(first: 1, orderBy: {field: CREATED_AT, direction: DESC})"`
			} `graphql:"repository(owner: $owner, name: $repo)"`
		}{}
		mockResponse.Organization.ProjectsV2.Nodes = []struct {
			Number    int
			CreatedAt time.Time
		}{{Number: 10, CreatedAt: time.Now()}}
		mockResponse.Repository.ProjectsV2.Nodes = []struct {
			Number    int
			CreatedAt time.Time
		}{{Number: 5, CreatedAt: time.Now().Add(-time.Hour)}}

		mockClient := &mockGQLClient{mockResponse: mockResponse}

		num, err := getLastProjectNumber(mockClient, "my-org", "my-repo")

		assert.NoError(t, err)
		assert.Equal(t, 10, num)
	})

	t.Run("Client returns an error", func(t *testing.T) {
		mockClient := &mockGQLClient{mockErr: fmt.Errorf("API rate limit exceeded")}

		_, err := getLastProjectNumber(mockClient, "my-org", "my-repo")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API rate limit exceeded")
	})

	t.Run("No projects found", func(t *testing.T) {
		mockClient := &mockGQLClient{mockResponse: &struct {
			Organization struct {
				ProjectsV2 struct{ Nodes []struct{ Number int; CreatedAt time.Time } } `graphql:"projectsV2(first: 1, orderBy: {field: CREATED_AT, direction: DESC})"`
			} `graphql:"organization(login: $owner)"`
			Repository   struct {
				ProjectsV2 struct{ Nodes []struct{ Number int; CreatedAt time.Time } } `graphql:"projectsV2(first: 1, orderBy: {field: CREATED_AT, direction: DESC})"`
			} `graphql:"repository(owner: $owner, name: $repo)"`
		}{}} // Empty response

		_, err := getLastProjectNumber(mockClient, "my-org", "my-repo")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no projects found")
	})
}

func TestFetchProjectData(t *testing.T) {
	t.Run("Successfully finds org project", func(t *testing.T) {
		mockOrgQueryResponse := &struct {
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
		mockOrgQueryResponse.Organization.ProjectV2 = &struct {
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

		mockClient := &mockGQLClient{mockResponse: mockOrgQueryResponse}

		items, title, err := fetchProjectData(mockClient, "my-org", "my-repo", 1, "")

		assert.NoError(t, err)
		assert.Equal(t, "My Org Project", title)
		assert.Len(t, items, 1)
		assert.Equal(t, "Test Draft", items[0].Content.DraftIssue.Title)
	})
}