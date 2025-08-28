package projects

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
)

type model struct {
	repoOwner     string
	repoName      string
	projectNumber int
	groupByField  string
	filter        itemFilter

	isLoading    bool
	spinner      spinner.Model
	table        table.Model   
	items        []ProjectItem 
	projectTitle string
	dividerRows map[int]bool
	err          error
}