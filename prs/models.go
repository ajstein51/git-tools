package prs

import "github.com/astein-peddi/git-tooling/models"

type Commit struct {
	Oid     string
	Message string
}

type branchScanResult struct {
	branchName string
	prs        []models.PR
	err        error
}