package projects

import (
	"github.com/charmbracelet/bubbles/table"
)

type model struct {
	repoOwner    string
	repoName     string
	projectTitle string
	groupByField string
	items        []ProjectItem

	table       table.Model
	dividerRows map[int]bool
}