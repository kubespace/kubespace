package git

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"time"
)

const (
	GithubType = "github"
)

type Secret struct {
	Type        string
	User        string
	Password    string
	PrivateKey  string
	AccessToken string
}

type Client interface {
	ListRepositories(ctx context.Context) ([]*Repository, error)
	ListRepoBranches(ctx context.Context, codeUrl string) ([]*Reference, error)
	GetBranchLatestCommit(ctx context.Context, codeUrl, branch string) (*Commit, error)
	Clone(repoDir string, isBare bool, options *git.CloneOptions) (*git.Repository, error)
	CreateTag(ctx context.Context, codeUrl, commitId, tagName string) error
}

func NewClient(gitType string, cloneUrl string, secret *Secret) (Client, error) {
	switch gitType {
	case GithubType:
		return NewGitHub(secret.AccessToken)
	}
	return nil, fmt.Errorf("git类型错误")
}

type Repository struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	CloneUrl string `json:"clone_url"`
}

type Commit struct {
	Branch     string    `json:"branch"`
	CommitId   string    `json:"commit_id"`
	Author     string    `json:"author"`
	Message    string    `json:"message"`
	CommitTime time.Time `json:"commit_time"`
}

type Reference struct {
	Name string `json:"name"`
	Ref  string `json:"ref"`
}

type PullRequest struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  string    `json:"username"`
}

type Git struct {
}
