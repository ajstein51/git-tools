package projects

import (
	"fmt"
	"sort"

	"github.com/astein-peddi/git-tooling/models"
	"github.com/astein-peddi/git-tooling/theme"
	"github.com/astein-peddi/git-tooling/utils"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cli/shurcooL-graphql"
	"github.com/pkg/browser"
)

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, fetchProjectItemsCmd(m.repoOwner, m.repoName, m.projectNumber, m.groupByField))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "enter":
			if len(m.items) == 0 {
				return m, nil
			}

			itemIndex := m.table.Cursor() - countDividersAbove(m.table.Cursor(), m.dividerRows)
			if itemIndex < 0 || itemIndex >= len(m.items) {
				return m, nil 
			}

			selectedItem := m.items[itemIndex]

			var url string
			switch selectedItem.Content.Typename {
				case "Issue":
					url = fmt.Sprintf("https://github.com/%s/%s/issues/%d", m.repoOwner, m.repoName, selectedItem.Content.Issue.Number)
				
				case "PullRequest":
					url = fmt.Sprintf("https://github.com/%s/%s/pull/%d", m.repoOwner, m.repoName, selectedItem.Content.PR.Number)
			}

			if url != "" {
				return m, openURLCmd(url)
			}

			return m, nil

		case "up", "k":
			m.table.MoveUp(1)
			for m.dividerRows[m.table.Cursor()] {
				if m.table.Cursor() == 0 {
					break
				}

				m.table.MoveUp(1)
			}

			return m, nil

		case "down", "j":
			m.table.MoveDown(1)
			for m.dividerRows[m.table.Cursor()] {
				if m.table.Cursor() == len(m.table.Rows())-1 {
					break
				}

				m.table.MoveDown(1)
			}

			return m, nil
		}

	case itemsLoadedMsg:
		m.isLoading = false
		m.projectTitle = msg.projectTitle

		processed := processProjectItems(msg.items, m.filter, m.groupByField)
		m.items = processed

		rows := []table.Row{}
		dividers := make(map[int]bool)
		var lastGroup string = "---no-group---"

		for _, item := range m.items {
			var itemType, numberStr, title, groupValue string
			switch item.Content.Typename {
				case "Issue":
					itemType = "Issue"
					numberStr = fmt.Sprintf("#%d", item.Content.Issue.Number)
					title = item.Content.Issue.Title

				case "PullRequest":
					itemType = "PR"
					numberStr = fmt.Sprintf("#%d", item.Content.PR.Number)
					title = item.Content.PR.Title

				case "DraftIssue":
					itemType = "Draft"
					numberStr = "-"
					title = item.Content.DraftIssue.Title
			}

			if m.groupByField != "" {
				groupValue = getFieldValue(item)
				
				if groupValue != lastGroup {
						dividerText := theme.DefaultTheme.Divider.Render(fmt.Sprintf("-- %s --", groupValue))
						rows = append(rows, table.Row{dividerText})
						dividers[len(rows)-1] = true

						lastGroup = groupValue
					}

					rows = append(rows, table.Row{itemType, numberStr, title, groupValue})
			} else {
				rows = append(rows, table.Row{itemType, numberStr, title})
			}
		}
		m.table.SetRows(rows)
		m.dividerRows = dividers

		if m.table.Cursor() > len(m.table.Rows())-1 {
			m.table.SetCursor(0)
		}
		
		if m.dividerRows[m.table.Cursor()] {
			m.table.MoveDown(1)
		}

		return m, nil

	case errorMsg:
		m.isLoading = false
		m.err = msg.err
		return m, tea.Quit

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)

		return m, cmd
	}

	m.table, cmd = m.table.Update(msg)

	return m, cmd
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nAn error occurred: %v\n\nPress q to quit.", m.err)
	}
	if m.isLoading {
		return fmt.Sprintf("\n %s Fetching items for project #%d...\n", m.spinner.View(), m.projectNumber)
	}

	var footer string
	helpText := "\n(↑/↓ to move (or vim motions), Enter to open, q to quit)"

	if len(m.items) > 0 {
		totalItems := len(m.items)

		currentItemIndex := m.table.Cursor() - countDividersAbove(m.table.Cursor(), m.dividerRows)
		
		currentItemNumber := currentItemIndex + 1

		paginationText := fmt.Sprintf("%d/%d", currentItemNumber, totalItems)

		footer = fmt.Sprintf("\n%s  %s", helpText, paginationText)
	} else {
		footer = "\n(q to quit)"
	}

	return lipgloss.NewStyle().Margin(1, 2).Render(
		fmt.Sprintf("Project #%d - %s\n\n%s%s",
			m.projectNumber,
			m.projectTitle,
			m.table.View(),
			theme.DefaultTheme.MutedText.Render(footer),
		),
	)
}

func initialModel(owner, repo string, projNum int, groupBy string, filter itemFilter) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = theme.DefaultTheme.Spinner

	columns := []table.Column{
		{Title: "Type", Width: 10},
		{Title: "Number", Width: 10},
		{Title: "Title", Width: 80},
	}

	if groupBy != "" {
		columns = append(columns, table.Column{Title: groupBy, Width: 20})
	}

	tbl := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	styles := table.DefaultStyles()
	styles.Header = theme.DefaultTheme.TableHeader
	styles.Selected = theme.DefaultTheme.SelectedListItem
	tbl.SetStyles(styles)

	return model{
		repoOwner:     owner,
		repoName:      repo,
		projectNumber: projNum,
		groupByField:  groupBy,
		filter:        filter,
		isLoading:     true,
		spinner:       s,
		table:         tbl,
		dividerRows:   make(map[int]bool),
	}
}

func openURLCmd(url string) tea.Cmd {
	return func() tea.Msg {
		go browser.OpenURL(url)

		return nil
	}
}

func countDividersAbove(cursor int, dividers map[int]bool) int {
	count := 0
	for i := 0; i < cursor; i++ {
		if dividers[i] {
			count++
		}
	}

	return count
}

func fetchProjectItemsCmd(owner, repo string, projNum int, groupBy string) tea.Cmd {
	return func() tea.Msg {
		client, err := utils.GetGhGraphQLClient()
		if err != nil {
			return errorMsg{err}
		}

		items, title, err := fetchProjectData(client, owner, repo, projNum, groupBy)
		if err != nil {
			return errorMsg{err}
		}

		return itemsLoadedMsg{items: items, projectTitle: title}
	}
}

func fetchProjectData(client models.GQLClient, owner, repo string, projectNumber int, groupByField string) ([]ProjectItem, string, error) {
	var allItems []ProjectItem
	var projectTitle string
	var foundProject bool

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

	if err := client.Query("OrgProjectItems", &orgQuery, orgVariables); err == nil && orgQuery.Organization.ProjectV2 != nil {
		foundProject = true
		projectTitle = orgQuery.Organization.ProjectV2.Title
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
					Title string
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

		if groupByField != "" {
			repoVariables["fieldName"] = graphql.String(groupByField)
		}

		if err := client.Query("RepoProjectItems", &repoQuery, repoVariables); err == nil && repoQuery.Repository.ProjectV2 != nil {
			foundProject = true
			projectTitle = repoQuery.Repository.ProjectV2.Title
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
		err := fmt.Errorf("failed to find project #%d in organization '%s' or repository '%s/%s'. Please check the project ID and your permissions", projectNumber, owner, owner, repo)
		return nil, "", err
	}

	return allItems, projectTitle, nil
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

			isIEmpty := fieldValueI == ""
			isJEmpty := fieldValueJ == ""

			if isIEmpty && !isJEmpty {
				return false
			}
			if !isIEmpty && isJEmpty {
				return true
			}

			return fieldValueI < fieldValueJ
		})
	} else {
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
