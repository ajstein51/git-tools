package projects

type PullRequestFragment struct {
	Number   int
	Title    string
	MergedAt *string
	ReviewRequests struct {
		Nodes []struct {
			RequestedReviewer struct {
				OnUser struct {
					Login string
				} `graphql:"... on User"`
			} `graphql:"requestedReviewer"`
		}
	} `graphql:"reviewRequests(first: 10)"`
}
