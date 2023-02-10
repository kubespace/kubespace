package git

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v50/github"
	"github.com/kubespace/kubespace/pkg/utils"
	"net/url"
	"strings"
)

type Github struct {
	client      *github.Client
	accessToken string
}

func NewGitHub(accessToken string) (*Github, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("not found params accessToken")
	}
	return &Github{
		client:      github.NewTokenClient(context.Background(), accessToken),
		accessToken: accessToken,
	}, nil
}

func (g *Github) ListRepositories(ctx context.Context) ([]*Repository, error) {
	res, _, err := g.client.Repositories.List(ctx, "", &github.RepositoryListOptions{})
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

func (g *Github) ListRepoRefs(ctx context.Context, codeUrl, matchRef string) ([]*Reference, error) {
	owner, repo, err := g.GetCodeOwnerRepo(codeUrl)
	if err != nil {
		return nil, err
	}
	repoRefs, _, err := g.client.Git.ListMatchingRefs(ctx, owner, repo, &github.ReferenceListOptions{
		Ref: matchRef,
	})
	if err != nil {
		return nil, err
	}
	var refs []*Reference
	for _, r := range repoRefs {
		refs = append(refs, &Reference{Name: r.GetRef(), Ref: r.GetRef()})
	}
	return refs, nil
}

func (g *Github) ListRepoBranches(ctx context.Context, codeUrl string) ([]*Reference, error) {
	return g.ListRepoRefs(ctx, codeUrl, "refs/heads")
}

func (g *Github) ListRepoPullRequests(ctx context.Context, codeUrl string) ([]*PullRequest, error) {
	owner, repo, err := g.GetCodeOwnerRepo(codeUrl)
	if err != nil {
		return nil, err
	}
	repoPRs, _, err := g.client.PullRequests.List(ctx, owner, repo, nil)
	if err != nil {
		return nil, err
	}
	var prs []*PullRequest
	for _, pr := range repoPRs {
		prs = append(prs, &PullRequest{
			Number:    pr.GetNumber(),
			Title:     pr.GetTitle(),
			State:     pr.GetState(),
			CreatedAt: pr.GetCreatedAt().In(utils.CSTZone),
			UpdatedAt: pr.GetUpdatedAt().In(utils.CSTZone),
			Username:  pr.GetUser().GetName(),
		})
	}
	return prs, nil
}

func (g *Github) CreateTag(ctx context.Context, codeUrl, commitId, tagName string) error {
	owner, repo, err := g.GetCodeOwnerRepo(codeUrl)
	if err != nil {
		return err
	}
	tag := "refs/tags/" + tagName
	_, _, err = g.client.Git.CreateRef(ctx, owner, repo, &github.Reference{
		Ref: &tag,
		Object: &github.GitObject{
			SHA: &commitId,
		},
	})
	return err
}

// GetCodeOwnerRepo 获取代码库的owner以及repo
// 如：https://github.com/test/testrepo.git -> test，testrepo
// git@github.com/test/testrepo.git -> test，testrepo
func (g *Github) GetCodeOwnerRepo(codeUrl string) (string, string, error) {
	if codeUrl[0:4] == "git@" {
		codeUrl = "https://" + codeUrl[4:]
	}
	u, err := url.Parse(codeUrl)
	if err != nil {
		return "", "", err
	}
	path := u.Path
	if path[0:1] == "/" {
		path = path[1:]
	}
	pathSplit := strings.SplitN(path, "/", 2)
	if len(pathSplit) != 2 {
		return "", "", fmt.Errorf("clone url=%s get owner and repo error: not found repo", codeUrl)
	}
	repoSplit := strings.Split(pathSplit[1], ".")
	return pathSplit[0], repoSplit[0], nil
}

func (g *Github) GetBranchLatestCommit(ctx context.Context, codeUrl, branch string) (*Commit, error) {
	owner, repo, err := g.GetCodeOwnerRepo(codeUrl)
	if err != nil {
		return nil, err
	}
	refName := "refs/heads/" + branch
	ref, _, err := g.client.Git.GetRef(ctx, owner, repo, refName)
	if err != nil {
		return nil, err
	}
	repoCommit, _, err := g.client.Repositories.GetCommit(ctx, owner, repo, ref.Object.GetSHA(), nil)
	if err != nil {
		return nil, err
	}
	commit := &Commit{
		Branch:     branch,
		Author:     repoCommit.Commit.Author.GetName(),
		Message:    repoCommit.Commit.GetMessage(),
		CommitTime: repoCommit.Commit.Author.Date.Time.In(utils.CSTZone),
	}
	return commit, nil
}

func (g *Github) Clone(repoDir string, isBare bool, options *git.CloneOptions) (*git.Repository, error) {
	auth := &http.BasicAuth{
		Username: "user",
		Password: g.accessToken,
	}
	options.InsecureSkipTLS = true
	options.Auth = auth
	return git.PlainClone(repoDir, isBare, options)
}
