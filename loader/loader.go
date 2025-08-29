package loader

import (
	"fmt"

	"github.com/astein-peddi/git-tooling/theme"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbletea"
)

type model struct {
	spinner  spinner.Model
	message  string
	resultCh chan Result
	result   Result
}

func waitForResultCmd(resultCh chan Result) tea.Cmd {
	return func() tea.Msg {
		return <-resultCh
	}
}

func InitialModel(message string, resultCh chan Result) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = theme.DefaultTheme.Spinner
	return model{
		spinner:  s,
		message:  message,
		resultCh: resultCh,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, waitForResultCmd(m.resultCh))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)

		return m, cmd

	case Result:
		m.result = msg

		return m, tea.Quit
	}

	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("\n %s %s...\n", m.spinner.View(), m.message)
}

func Run(message string, task func() (any, error)) (any, error) {
	resultCh := make(chan Result, 1)

	go func() {
		data, err := task()
		resultCh <- Result{Data: data, Err: err}
	}()

	p := tea.NewProgram(InitialModel(message, resultCh))
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("error running loader: %w", err)
	}

	result := finalModel.(model).result
	return result.Data, result.Err
}
