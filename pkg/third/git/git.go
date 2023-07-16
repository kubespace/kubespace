package git

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	sshgit "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog/v2"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

type Client interface {
	Auth() (transport.AuthMethod, error)
	ListRepositories(ctx context.Context) ([]*Repository, error)
	ListRepoBranches(ctx context.Context, codeUrl string) ([]*Reference, error)
	GetBranchLatestCommit(ctx context.Context, codeUrl, branch string) (*Commit, error)
	Clone(ctx context.Context, repoDir string, isBare bool, options *git.CloneOptions) (*git.Repository, error)
	CreateTag(ctx context.Context, codeUrl, commitId, tagName string) error
}

func NewClient(gitType string, apiUrl string, secret *types.Secret) (Client, error) {
	switch gitType {
	case types.WorkspaceCodeTypeGitHub:
		return NewGitHub(secret.AccessToken)
	case types.WorkspaceCodeTypeGitLab:
		return NewGitLab(apiUrl, secret.AccessToken)
	case types.WorkspaceCodeTypeGitee:
		return NewGitee(secret.AccessToken)
	case types.WorkspaceCodeTypeHttps, types.WorkspaceCodeTypeGit:
		return NewGit(secret), nil
	}
	return nil, fmt.Errorf("git类型错误")
}

// GetCodeOwnerRepo 获取代码库的owner以及repo
// 如：https://github.com/test/testrepo.git -> test，testrepo
// git@github.com/test/testrepo.git -> test，testrepo
func GetCodeOwnerRepo(codeUrl string) (string, string, error) {
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
	Name     string `json:"name"`
	Ref      string `json:"ref"`
	CommitId string `json:"commit_id"`
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
	secret *types.Secret
}

func NewGit(secret *types.Secret) *Git {
	return &Git{secret: secret}
}

func (g *Git) ListRepositories(ctx context.Context) ([]*Repository, error) {
	return nil, fmt.Errorf("not implement list git repositories")
}

func (g *Git) Auth() (transport.AuthMethod, error) {
	var auth transport.AuthMethod
	switch g.secret.Type {
	case types.SettingsSecretTypeKey:
		if g.secret.PrivateKey == "" {
			return nil, fmt.Errorf("代码私钥为空")
		}
		privateKey, err := sshgit.NewPublicKeys("git", []byte(g.secret.PrivateKey), "")
		if err != nil {
			return nil, fmt.Errorf("生成代码密钥失败：" + err.Error())
		}
		privateKey.HostKeyCallbackHelper = sshgit.HostKeyCallbackHelper{
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		auth = privateKey
	case types.SettingsSecretTypePassword:
		auth = &http.BasicAuth{
			Username: g.secret.User,
			Password: g.secret.Password,
		}
	}
	return auth, nil
}

func (g *Git) ListRepoBranches(ctx context.Context, codeUrl string) ([]*Reference, error) {
	auth, err := g.Auth()
	if err != nil {
		return nil, err
	}
	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{codeUrl},
	})
	remoteRefs, err := rem.ListContext(ctx, &git.ListOptions{Auth: auth, InsecureSkipTLS: true})
	if err != nil {
		return nil, fmt.Errorf("获取代码远程分支失败：" + err.Error())
	}
	var refs []*Reference
	for _, ref := range remoteRefs {
		if ref.Name().IsBranch() {
			refs = append(refs, &Reference{
				Name:     ref.Name().Short(),
				Ref:      ref.Name().String(),
				CommitId: ref.Hash().String(),
			})
		}
	}
	return refs, nil
}

func (g *Git) GetBranchLatestCommit(ctx context.Context, codeUrl, branch string) (*Commit, error) {
	auth, err := g.Auth()
	if err != nil {
		return nil, err
	}
	uuid := utils.ShortUUID()
	refName := "refs/heads/" + branch
	ref, err := git.PlainCloneContext(ctx, "/tmp/"+uuid, true, &git.CloneOptions{
		Auth:            auth,
		URL:             codeUrl,
		Progress:        os.Stdout,
		ReferenceName:   plumbing.ReferenceName(refName),
		SingleBranch:    true,
		Depth:           1,
		NoCheckout:      true,
		InsecureSkipTLS: true,
	})
	if err != nil {
		klog.Errorf("git clone %s error: %v", codeUrl, err)
		return nil, err
	}
	defer os.RemoveAll("/tmp/" + uuid)
	commits, err := ref.Log(&git.LogOptions{})
	if err != nil {
		klog.Errorf("git log %s error: %v", codeUrl, err)
		return nil, err
	}
	commit, err := commits.Next()
	if err != nil {
		klog.Errorf("git log %s error: %v", codeUrl, err)
		return nil, err
	}
	return &Commit{
		Branch:     branch,
		CommitId:   commit.Hash.String(),
		Author:     commit.Author.Name,
		Message:    commit.Message,
		CommitTime: commit.Author.When,
	}, nil
}

func (g *Git) Clone(ctx context.Context, repoDir string, isBare bool, options *git.CloneOptions) (*git.Repository, error) {
	auth, err := g.Auth()
	if err != nil {
		return nil, err
	}
	options.InsecureSkipTLS = true
	options.Auth = auth
	return git.PlainCloneContext(ctx, repoDir, isBare, options)
}

func (g *Git) CreateTag(ctx context.Context, codeUrl, commitId, tagName string) error {
	codeDir := path.Join("/tmp", utils.ShortUUID())
	repo, err := g.Clone(ctx, codeDir, false, &git.CloneOptions{
		URL:      codeUrl,
		Progress: os.Stdout,
	})
	if err != nil {
		klog.Errorf("create tag %s clone %s error: %v", tagName, codeUrl, err)
		return fmt.Errorf("git clone %s error: %v", codeUrl, err)
	}
	defer os.RemoveAll(codeDir)
	w, err := repo.Worktree()
	if err != nil {
		klog.Errorf("create tag %s clone %s error: %v", tagName, codeUrl, err)
		return fmt.Errorf("git clone %s error: %v", codeUrl, err)
	}
	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(commitId),
	})
	if err != nil {
		klog.Errorf("git checkout %s error: %v", commitId, err)
		return fmt.Errorf("git url %s checkout %s error: %v", codeUrl, commitId, err)
	}
	_, err = repo.CreateTag(tagName, plumbing.NewHash(commitId), &git.CreateTagOptions{
		Message: tagName,
		Tagger: &object.Signature{
			Name:  "kubespace",
			Email: "kubespace@kubespace.cn",
			When:  time.Now(),
		},
	})
	if err != nil {
		if !errors.Is(err, git.ErrTagExists) {
			return fmt.Errorf("git tag error: %s", err.Error())
		}
	}
	auth, err := g.Auth()
	if err != nil {
		return err
	}
	po := &git.PushOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/tags/*:refs/tags/*")},
		Auth:       auth,
	}
	err = repo.PushContext(ctx, po)
	if err != nil {
		return fmt.Errorf("push tag %s error: %s", tagName, err.Error())
	}

	return nil
}
