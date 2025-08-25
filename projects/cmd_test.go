package projects

// This was straight vibe coded. Tests may or may not be comprehensive.

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cli/shurcooL-graphql"
	"github.com/stretchr/testify/assert"
)

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
			// assert.Equal doesn't work well with slices of structs, so we check len and contents.
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
		draft.withCustomField(""), // Empty custom field
	}

	// Define filters from the command setup
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
		expectedTitles []string // We check titles to verify filtering and sorting
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
				"[In Progress] PR #102: Open PR",
				// Expect a blank line between groups
				"",
				"[Todo] Issue #1: My Issue",
				"[Todo] Draft: A draft",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			displayProjectItems(items, 42, "Test Project", tc.groupByField)

			// Restore stdout
			w.Close()
			os.Stdout = old
			var buf bytes.Buffer
			io.Copy(&buf, r)

			output := strings.TrimSpace(buf.String())
			expectedOutput := strings.Join(tc.expectedLines, "\n")

			// Normalize line endings for cross-platform compatibility
			normalizedOutput := strings.ReplaceAll(output, "\r\n", "\n")

			assert.Equal(t, expectedOutput, normalizedOutput)
		})
	}
}

// --- Tests for API-dependent functions ---

func TestGetLastProjectNumber(t *testing.T) {
	t.Skip(`Skipping due to direct dependency on utils.GetGhGraphQLClient(). 
To test this, the function should be refactored to accept a GraphQL client interface, 
allowing a mock client to be injected during tests.`)
}

func TestFetchProjectData(t *testing.T) {
	t.Skip(`Skipping due to direct dependency on an api.GraphQLClient instance.
To test this, this function should accept a client interface.
Test cases would include:
- Finding a project in the organization.
- Finding a project in the repository.
- Handling pagination correctly.
- Returning an error when the project is not found in either.`)
}