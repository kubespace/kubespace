package git

import "context"

const (
	GithubType = "github"
)

type Interface interface {
	ListRepositories(ctx context.Context) ([]*Repository, error)
}

func NewClient(gitType string, apiUrl string, accessToken string) Interface {
	switch gitType {
	case GithubType:
		return NewGitHub(accessToken)
	}
	return nil
}

type Repository struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	CloneUrl string `json:"clone_url"`
}
