package prs

type PR struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
}

type Commit struct {
	Oid     string
	Message string
}

type branchScanResult struct {
	branchName string
	prs        []PR
	err        error
}