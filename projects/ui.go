package projects

import (
	"fmt"

	"github.com/astein-peddi/git-tooling/theme"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/browser"
)

func (m Model) Init() tea.Cmd {
	return tea.WindowSize()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.table = setupTable(msg.Width, m.items, m.groupByField, m.dividerRows)
			return m, nil

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
			}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.table.Columns() == nil {
		return "Initializing..."
	}

	var footer string
	helpText := "(↑/↓ to move or vim motions, Enter to open, q to quit)"

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
		fmt.Sprintf("%s\n\n%s\n%s",
			m.projectTitle,
			m.table.View(),
			theme.DefaultTheme.MutedText.Render(footer),
		),
	)
}

func initialModel(owner, repo, title string, items []ProjectItem, groupBy string) Model {
	return Model {
		repoOwner:    owner,
		repoName:     repo,
		projectTitle: title,
		groupByField: groupBy,
		items:        items,
		dividerRows:  make(map[int]bool),
	}
}

func setupTable(termWidth int, items []ProjectItem, groupBy string, dividers map[int]bool) table.Model {
	typeWidth := 10
	numWidth := 10
	groupWidth := 20
	padding := 8

	titleWidth := termWidth - typeWidth - numWidth - padding
	if groupBy != "" {
		titleWidth -= groupWidth
	}
	if titleWidth < 20 {
		titleWidth = 20
	}

	columns := []table.Column{
		{Title: "Type", Width: typeWidth},
		{Title: "Number", Width: numWidth},
		{Title: "Title", Width: titleWidth},
	}
	if groupBy != "" {
		columns = append(columns, table.Column{Title: groupBy, Width: groupWidth})
	}

	tbl := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(30), 
	)
	styles := table.DefaultStyles()
	styles.Header = theme.DefaultTheme.TableHeader
	styles.Selected = theme.DefaultTheme.SelectedListItem
	tbl.SetStyles(styles)

	rows := []table.Row{}
	var lastGroup string = "---no-group---"
	for k := range dividers {
		delete(dividers, k)
	}
	for _, item := range items {
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

		if groupBy != "" {
			groupValue = getFieldValue(item)
			if groupValue != lastGroup {
				dividerText := theme.DefaultTheme.Divider.Render(fmt.Sprintf("-- %s --", ""))
				rows = append(rows, table.Row{dividerText})
				dividers[len(rows)-1] = true
				lastGroup = groupValue
			}

			rows = append(rows, table.Row{itemType, numberStr, title, groupValue})
		} else {
			rows = append(rows, table.Row{itemType, numberStr, title})
		}
	}

	tbl.SetRows(rows)
	return tbl
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