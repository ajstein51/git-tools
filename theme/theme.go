package theme

import "github.com/charmbracelet/lipgloss"

type AppTheme struct {
	SelectedListItem lipgloss.Style
	Spinner          lipgloss.Style
	TableHeader      lipgloss.Style
	Divider          lipgloss.Style
	MutedText        lipgloss.Style
}

var DefaultTheme = NordTheme

var OriginalTheme = AppTheme{
	SelectedListItem: lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")),
	Spinner:          lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
	TableHeader:      lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).BorderBottom(true),
	Divider:          lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("240")),
	MutedText:        lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
}

// MonokaiTheme is inspired by the popular Monokai editor theme.
var MonokaiTheme = AppTheme{
	SelectedListItem: lipgloss.NewStyle().Foreground(lipgloss.Color("#272822")).Background(lipgloss.Color("#F92672")), // Black on Pink
	Spinner:          lipgloss.NewStyle().Foreground(lipgloss.Color("#AE81FF")), // Purple
	TableHeader:      lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("244")).BorderBottom(true),
	Divider:          lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A6E22E")), // Green
	MutedText:        lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
}

// GruvboxTheme uses a retro, warm color palette.
var GruvboxTheme = AppTheme{
	SelectedListItem: lipgloss.NewStyle().Foreground(lipgloss.Color("#282828")).Background(lipgloss.Color("#FE8019")), // Dark on Orange
	Spinner:          lipgloss.NewStyle().Foreground(lipgloss.Color("#8EC07C")), // Green
	TableHeader:      lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("245")).BorderBottom(true),
	Divider:          lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#458588")), // Blue/Aqua
	MutedText:        lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
}

// NordTheme is a cool, elegant, arctic-inspired theme.
var NordTheme = AppTheme{
	SelectedListItem: lipgloss.NewStyle().Foreground(lipgloss.Color("#ECEFF4")).Background(lipgloss.Color("#5E81AC")), // Light on Blue
	Spinner:          lipgloss.NewStyle().Foreground(lipgloss.Color("#B48EAD")), // Purple
	TableHeader:      lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#4C566A")).BorderBottom(true),
	Divider:          lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#88C0D0")), // Light Blue
	MutedText:        lipgloss.NewStyle().Foreground(lipgloss.Color("#4C566A")),
}

// MonochromeTheme is a simple, high-contrast theme that works on all terminals.
var MonochromeTheme = AppTheme{
	SelectedListItem: lipgloss.NewStyle().Reverse(true), // Inverts foreground and background
	Spinner:          lipgloss.NewStyle().Bold(true),
	TableHeader:      lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).Bold(true),
	Divider:          lipgloss.NewStyle().Faint(true), // Dim text
	MutedText:        lipgloss.NewStyle().Faint(true),
}