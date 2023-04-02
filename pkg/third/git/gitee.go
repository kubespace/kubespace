package git

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/kubespace/kubespace/pkg/third/httpclient"
	"github.com/kubespace/kubespace/pkg/utils"
	"time"
)

const (
	defaultBaseURL = "https://gitee.com/"
	apiVersionPath = "api/v5/"
)

type Gitee struct {
	httpClient  *httpclient.HttpClient
	accessToken string
}

func NewGitee(accessToken string) (*Gitee, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("not found gitee access token params")
	}
	httpCli, err := httpclient.NewHttpClient(defaultBaseURL + apiVersionPath)
	if err != nil {
		return nil, err
	}
	return &Gitee{
		httpClient:  httpCli,
		accessToken: accessToken,
	}, nil
}

type GiteeRepositoryListQuery struct {
	AccessToken string `json:"access_token" url:"access_token"`
}

type GiteeRepository struct {
	Id       int    `json:"id"`
	Path     string `json:"path"`
	FullName string `json:"full_name"`
	HtmlUrl  string `json:"html_url"`
}

func (g *Gitee) Auth() (transport.AuthMethod, error) {
	query := &GiteeRepositoryListQuery{AccessToken: g.accessToken}
	var user GiteeUser
	if _, err := g.httpClient.Get("user", query, &user, httpclient.RequestOptions{}); err != nil {
		return nil, err
	}

	return &http.BasicAuth{
		Username: user.Login,
		Password: g.accessToken,
	}, nil
}

func (g *Gitee) ListRepositories(ctx context.Context) ([]*Repository, error) {
	query := &GiteeRepositoryListQuery{AccessToken: g.accessToken}
	var giteeRepos []*GiteeRepository
	if _, err := g.httpClient.Get("user/repos", query, &giteeRepos, httpclient.RequestOptions{Context: ctx}); err != nil {
		return nil, err
	}
	var repos []*Repository
	for _, r := range giteeRepos {
		repos = append(repos, &Repository{
			Name:     r.Path,
			FullName: r.FullName,
			CloneUrl: r.HtmlUrl,
		})
	}
	return repos, nil
}

type GiteeRepoBranchListQuery struct {
	AccessToken string `json:"access_token" url:"access_token"`
	Owner       string `json:"owner" url:"owner"`
	Repo        string `json:"repo" url:"repo"`
}

type GiteeBranchCommit struct {
	SHA string `json:"sha"`
	Url string `json:"url"`
}

type GiteeRepoBranch struct {
	Commit       *GiteeBranchCommit `json:"commit"`
	Name         string             `json:"name"`
	Protected    bool               `json:"protected"`
	ProtectedUrl string             `json:"protected_url"`
}

func (g *Gitee) ListRepoBranches(ctx context.Context, codeUrl string) ([]*Reference, error) {
	owner, repo, err := GetCodeOwnerRepo(codeUrl)
	if err != nil {
		return nil, err
	}
	query := &GiteeRepositoryListQuery{AccessToken: g.accessToken}
	path := fmt.Sprintf("repos/%s/%s/branches", owner, repo)
	var giteeBranches []*GiteeRepoBranch
	if _, err = g.httpClient.Get(path, query, &giteeBranches, httpclient.RequestOptions{Context: ctx}); err != nil {
		return nil, err
	}
	var refs []*Reference
	for _, b := range giteeBranches {
		refs = append(refs, &Reference{
			Name:     b.Name,
			Ref:      b.Name,
			CommitId: b.Commit.SHA,
		})
	}
	return refs, nil
}

type GiteeListPullRequests struct {
	AccessToken string `json:"access_token" url:"access_token"`
	State       string `json:"state" url:"state"`
}

type GiteePullRequest struct {
	Id        int       `json:"id"`
	Number    int       `json:"int"`
	State     string    `json:"state"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (g *Gitee) ListRepoPullRequests(ctx context.Context, codeUrl string) ([]*PullRequest, error) {
	owner, repo, err := GetCodeOwnerRepo(codeUrl)
	if err != nil {
		return nil, err
	}
	query := &GiteeListPullRequests{AccessToken: g.accessToken, State: "open"}
	path := fmt.Sprintf("repos/%s/%s/pulls", owner, repo)
	var giteePrs []*GiteePullRequest
	if _, err = g.httpClient.Get(path, query, &giteePrs, httpclient.RequestOptions{Context: ctx}); err != nil {
		return nil, err
	}
	var prs []*PullRequest
	for _, pr := range giteePrs {
		prs = append(prs, &PullRequest{
			Number:    pr.Number,
			Title:     pr.Title,
			State:     pr.State,
			CreatedAt: pr.CreatedAt,
			UpdatedAt: pr.UpdatedAt,
			Username:  "",
		})
	}
	return prs, nil
}

type GiteeCreateTagRequest struct {
	AccessToken string `json:"access_token"`
	Refs        string `json:"refs"`
	TagName     string `json:"tag_name"`
}

func (g *Gitee) CreateTag(ctx context.Context, codeUrl, commitId, tagName string) error {
	owner, repo, err := GetCodeOwnerRepo(codeUrl)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("repos/%s/%s/tags", owner, repo)
	req := &GiteeCreateTagRequest{
		AccessToken: g.accessToken,
		Refs:        commitId,
		TagName:     tagName,
	}
	if _, err = g.httpClient.Post(path, req, nil, httpclient.RequestOptions{Context: ctx}); err != nil {
		return err
	}
	return nil
}

type GiteeCommit struct {
	Url     string `json:"url"`
	SHA     string `json:"sha"`
	HtmlUrl string `json:"html_url"`
	Commit  struct {
		Author struct {
			Name  string    `json:"name"`
			Date  time.Time `json:"date"`
			Email string    `json:"email"`
		} `json:"author"`
		Message string `json:"message"`
	} `json:"commit"`
}

func (g *Gitee) GetBranchLatestCommit(ctx context.Context, codeUrl, branch string) (*Commit, error) {
	owner, repo, err := GetCodeOwnerRepo(codeUrl)
	if err != nil {
		return nil, err
	}
	query := &GiteeRepositoryListQuery{AccessToken: g.accessToken}
	path := fmt.Sprintf("repos/%s/%s/commits/%s", owner, repo, branch)
	var repoCommit GiteeCommit
	if _, err = g.httpClient.Get(path, query, &repoCommit, httpclient.RequestOptions{Context: ctx}); err != nil {
		return nil, err
	}
	commit := &Commit{
		Branch:     branch,
		Author:     repoCommit.Commit.Author.Name,
		Message:    repoCommit.Commit.Message,
		CommitTime: repoCommit.Commit.Author.Date.In(utils.CSTZone),
	}
	return commit, nil
}

type GiteeUser struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Login string `json:"login"`
}

func (g *Gitee) Clone(ctx context.Context, repoDir string, isBare bool, options *git.CloneOptions) (*git.Repository, error) {
	query := &GiteeRepositoryListQuery{AccessToken: g.accessToken}
	var user GiteeUser
	if _, err := g.httpClient.Get("user", query, &user, httpclient.RequestOptions{Context: ctx}); err != nil {
		return nil, err
	}

	// gitee这个地方需要使用Login字段进行认证
	auth := &http.BasicAuth{
		Username: user.Login,
		Password: g.accessToken,
	}

	options.InsecureSkipTLS = true
	options.Auth = auth
	return git.PlainCloneContext(ctx, repoDir, isBare, options)
}
