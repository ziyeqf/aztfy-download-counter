package datasource

import (
	"context"

	"github.com/google/go-github/v50/github"
)

const GithubPerPage = 20
const RepoOwner = "Azure"
const RepoName = "aztfexport"

func FetchGitHubDownloadCount(ctx context.Context) ([]*github.RepositoryRelease, error) {
	client := github.NewClient(nil)
	result := make([]*github.RepositoryRelease, 0)

	itemCount := 0
	for i := 1; itemCount < GithubPerPage; i++ {
		itemCount = 0

		opt := &github.ListOptions{
			Page:    i,
			PerPage: GithubPerPage,
		}

		releases, _, err := client.Repositories.ListReleases(ctx, RepoOwner, RepoName, opt)
		if err != nil {
			return nil, err
		}

		result = append(result, releases...)
		itemCount = len(releases)
	}

	return result, nil
}
