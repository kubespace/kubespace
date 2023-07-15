package git

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v50/github"
	"github.com/kubespace/kubespace/pkg/utils"
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

func (g *Github) Auth() (transport.AuthMethod, error) {
	return &http.BasicAuth{
		Username: "user",
		Password: g.accessToken,
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
	owner, repo, err := GetCodeOwnerRepo(codeUrl)
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
		refs = append(refs, &Reference{
			Name:     r.GetRef(),
			Ref:      r.GetRef(),
			CommitId: r.Object.GetSHA(),
		})
	}
	return refs, nil
}

func (g *Github) ListRepoBranches(ctx context.Context, codeUrl string) ([]*Reference, error) {
	if refs, err := g.ListRepoRefs(ctx, codeUrl, "refs/heads"); err != nil {
		return nil, err
	} else {
		for _, r := range refs {
			n, f := strings.CutPrefix(r.Name, "refs/heads/")
			if f {
				r.Name = n
			}
		}
		return refs, nil
	}
}

func (g *Github) ListRepoPullRequests(ctx context.Context, codeUrl string) ([]*PullRequest, error) {
	owner, repo, err := GetCodeOwnerRepo(codeUrl)
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
	owner, repo, err := GetCodeOwnerRepo(codeUrl)
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

func (g *Github) GetBranchLatestCommit(ctx context.Context, codeUrl, branch string) (*Commit, error) {
	owner, repo, err := GetCodeOwnerRepo(codeUrl)
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
		CommitId:   repoCommit.GetSHA(),
		Author:     repoCommit.Commit.Author.GetName(),
		Message:    repoCommit.Commit.GetMessage(),
		CommitTime: repoCommit.Commit.Author.Date.Time.In(utils.CSTZone),
	}
	return commit, nil
}

func (g *Github) Clone(ctx context.Context, repoDir string, isBare bool, options *git.CloneOptions) (*git.Repository, error) {
	auth := &http.BasicAuth{
		Username: "user",
		Password: g.accessToken,
	}
	options.InsecureSkipTLS = true
	options.Auth = auth
	return git.PlainCloneContext(ctx, repoDir, isBare, options)
}
