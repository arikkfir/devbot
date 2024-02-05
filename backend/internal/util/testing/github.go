package testing

import (
	"context"
	"github.com/google/go-github/v56/github"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func BranchHasName(name string) types.GomegaMatcher {
	return WithTransform(func(b *github.Branch) string { return b.GetName() }, Equal(name))
}

func ListGitHubBranches(ctx context.Context, gh *github.Client, owner, repoName string, targetBranches *[]*github.Branch) {
	var branches []*github.Branch
	branchesListOptions := &github.BranchListOptions{}
	for {
		branchesList, response, err := gh.Repositories.ListBranches(ctx, owner, repoName, branchesListOptions)
		Expect(err).NotTo(HaveOccurred())
		for _, branch := range branchesList {
			branches = append(branches, branch)
		}
		if response.NextPage == 0 {
			break
		}
		branchesListOptions.Page = response.NextPage
	}
	*targetBranches = branches
}
