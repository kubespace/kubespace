package git

import (
	"context"
	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
)

type Github struct {
	client *github.Client
}

func NewGitHub(accessToken string) *Github {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	return &Github{
		client: github.NewClient(tc),
	}
}

func (g *Github) ListRepositories(ctx context.Context) ([]*Repository, error) {
	res, _, err := g.client.Repositories.List(ctx, "", nil)
	if err != nil {
		return nil, err
	}
	var repos []*Repository
	for _, r := range res {
		repos = append(repos, &Repository{
			Name:     *r.Name,
			FullName: *r.FullName,
			CloneUrl: *r.CloneURL,
		})
	}
	return repos, nil
}
