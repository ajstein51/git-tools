package prs

type PR struct {
	Number      int
	Title       string
	URL         string
	MergeCommit struct {
		Oid string
	}
}