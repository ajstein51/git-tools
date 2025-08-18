package prs

type PR struct {
	Number      string `json:"number"`
	Title       string `json:"title"`
	ShortCommit string `json:"short_commit"`
	URL         string `json:"url"`
}
