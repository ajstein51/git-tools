package prs

import (
	"fmt"

	"github.com/astein-peddi/git-tooling/theme"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	branchA string
	branchB string
	prs     []PR
	table   table.Model
}

func initialModel(branchA, branchB string, prs []PR) model {
	return model{
		branchA: branchA,
		branchB: branchB,
		prs:     prs,
	}
}

func (m model) Init() tea.Cmd {
	return tea.WindowSize()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.table = setupTable(msg.Width, m.prs)
			return m, nil

		case tea.KeyMsg:
			switch msg.String() {
				case "q", "ctrl+c":
					return m, tea.Quit
			}
	}
	
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// In prs/ui.go

func (m model) View() string {
	if m.table.Columns() == nil {
		return "Initializing..."
	}

	var header, footer string
	
	header = fmt.Sprintf("PRs merged into '%s' but not in '%s'\n\n", m.branchA, m.branchB)
	
	helpText := "(q to quit)"
	if len(m.prs) > 0 {
		helpText = "(↑/↓ to move or Vim Motions, q to quit)"
		paginationText := fmt.Sprintf("%d/%d", m.table.Cursor()+1, len(m.prs))
		footer = fmt.Sprintf("\n\n%s  %s", helpText, paginationText)
	} else {
		footer = fmt.Sprintf("\n\n%s", helpText)
	}
	
	var body string
	if len(m.prs) > 0 {
		body = m.table.View()
	} else {
		body = "No differences found. (q to quit)"
	}
	
	return lipgloss.NewStyle().Margin(1, 2).Render(
		header +
		body +
		theme.DefaultTheme.MutedText.Render(footer),
	)
}

func setupTable(termWidth int, prs []PR) table.Model {
	numWidth := 10
	padding := 8
	titleWidth := max(termWidth - numWidth - padding, 20)

	columns := []table.Column{
		{Title: "Number", Width: numWidth},
		{Title: "Title", Width: titleWidth},
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

	rows := []table.Row{}
	for _, pr := range prs {
		rows = append(rows, table.Row{fmt.Sprintf("#%d", pr.Number), pr.Title})
	}
	tbl.SetRows(rows)
	return tbl
}