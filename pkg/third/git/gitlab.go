package git

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/xanzy/go-gitlab"
)

type Gitlab struct {
	client      *gitlab.Client
	accessToken string
}

func NewGitLab(apiUrl, accessToken string) (*Gitlab, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("not found params accessToken")
	}
	client, err := gitlab.NewClient(accessToken, gitlab.WithBaseURL(apiUrl))
	if err != nil {
		return nil, err
	}
	return &Gitlab{
		client:      client,
		accessToken: accessToken,
	}, nil
}

func (g *Gitlab) Auth() (transport.AuthMethod, error) {
	return &http.BasicAuth{
		Username: "user",
		Password: g.accessToken,
	}, nil
}

func (g *Gitlab) ListRepositories(ctx context.Context) ([]*Repository, error) {
	owned := true
	res, _, err := g.client.Projects.ListProjects(&gitlab.ListProjectsOptions{Owned: &owned}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	var repos []*Repository
	for _, r := range res {
		repos = append(repos, &Repository{
			Name:     r.Path,
			FullName: r.PathWithNamespace,
			CloneUrl: r.HTTPURLToRepo,
		})
	}
	return repos, nil
}

func (g *Gitlab) GetPID(codeUrl string) (string, error) {
	owner, repo, err := GetCodeOwnerRepo(codeUrl)
	if err != nil {
		return "", err
	}
	return owner + "/" + repo, nil
}

func (g *Gitlab) ListRepoBranches(ctx context.Context, codeUrl string) ([]*Reference, error) {
	pid, err := g.GetPID(codeUrl)
	if err != nil {
		return nil, err
	}
	repoBranches, _, err := g.client.Branches.ListBranches(pid, nil, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	var branches []*Reference
	for _, b := range repoBranches {
		branches = append(branches, &Reference{
			Name:     b.Name,
			Ref:      b.Name,
			CommitId: b.Commit.ID,
		})
	}
	return branches, nil
}

func (g *Gitlab) ListRepoPullRequests(ctx context.Context, codeUrl string) ([]*PullRequest, error) {
	return nil, nil
}

func (g *Gitlab) CreateTag(ctx context.Context, codeUrl, commitId, tagName string) error {
	pid, err := g.GetPID(codeUrl)
	if err != nil {
		return err
	}
	_, _, err = g.client.Tags.CreateTag(pid,
		&gitlab.CreateTagOptions{TagName: &tagName, Ref: &commitId}, gitlab.WithContext(ctx))
	return err
}

func (g *Gitlab) GetBranchLatestCommit(ctx context.Context, codeUrl, branch string) (*Commit, error) {
	pid, err := g.GetPID(codeUrl)
	if err != nil {
		return nil, err
	}
	commit, _, err := g.client.Commits.GetCommit(pid, branch, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return &Commit{
		Branch:     branch,
		CommitId:   commit.ID,
		Author:     commit.AuthorName,
		Message:    commit.Message,
		CommitTime: commit.CommittedDate.In(utils.CSTZone),
	}, nil
}

func (g *Gitlab) Clone(ctx context.Context, repoDir string, isBare bool, options *git.CloneOptions) (*git.Repository, error) {
	auth := &http.BasicAuth{
		Username: "user",
		Password: g.accessToken,
	}
	options.InsecureSkipTLS = true
	options.Auth = auth
	return git.PlainCloneContext(ctx, repoDir, isBare, options)
}
