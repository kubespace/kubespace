package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/service/pipeline/schemas"
	utilgit "github.com/kubespace/kubespace/pkg/third/git"
	"github.com/kubespace/kubespace/pkg/third/httpclient"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const PipelineRunAddReleaseVersionUri = "/api/v1/spacelet/pipeline/add_release"

type ReleaserPlugin struct {
	client *httpclient.HttpClient
}

func NewReleasePlugin(client *httpclient.HttpClient) *ReleaserPlugin {
	return &ReleaserPlugin{client: client}
}

func (b ReleaserPlugin) Executor(params *ExecutorParams) (Executor, error) {
	return newReleaserExecutor(params, b.client)
}

type releaseParams struct {
	JobId       uint `json:"job_id"`
	WorkspaceId uint `json:"workspace_id"`

	CodeUrl      string        `json:"code_url"`
	CodeApiUrl   string        `json:"code_api_url"`
	CodeType     string        `json:"code_type"`
	CodeBranch   string        `json:"code_branch"`
	CodeCommitId string        `json:"code_commit_id"`
	CodeSecret   *types.Secret `json:"code_secret"`

	ImageBuildRegistry types.ImageRegistry `json:"image_registry"`

	Version string `json:"version"`
	Images  string `json:"images"`
}

type ReleaserPluginResult struct {
	Version string `json:"version"`
	Images  string `json:"images"`
}

type releaserExecutor struct {
	Logger
	client     *httpclient.HttpClient
	Params     *releaseParams
	CodeDir    string
	Result     *ReleaserPluginResult
	Images     map[string]string
	ctx        context.Context
	cancelFunc context.CancelFunc
	canceled   bool
}

func newReleaserExecutor(pluginParams *ExecutorParams, client *httpclient.HttpClient) (*releaserExecutor, error) {
	var params releaseParams
	if err := utils.ConvertTypeByJson(pluginParams.Params, &params); err != nil {
		return nil, err
	}
	if params.Version == "" {
		return nil, fmt.Errorf("发布版本号为空")
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	plugin := &releaserExecutor{
		client: client,
		Logger: pluginParams.Logger,
		Params: &params,
		Result: &ReleaserPluginResult{
			Version: params.Version,
			Images:  "",
		},
		ctx:        ctx,
		cancelFunc: cancelFunc,
		Images:     make(map[string]string),
	}
	codeDir := utils.GetCodeRepoName(params.CodeUrl)
	if codeDir == "" {
		klog.Errorf("job=%d get empty code repo name", params.JobId)
		return nil, fmt.Errorf("get empty code repo name")
	}
	plugin.CodeDir, _ = filepath.Abs(filepath.Join(pluginParams.RootDir, codeDir))

	return plugin, nil
}

func (r *releaserExecutor) Execute() (interface{}, error) {
	var addVersionResp utils.Response
	addVersionParams := &schemas.AddReleaseVersionParams{
		WorkspaceId: r.Params.WorkspaceId,
		JobId:       r.Params.JobId,
		Version:     r.Params.Version,
	}
	if _, err := r.client.Post(PipelineRunAddReleaseVersionUri, addVersionParams, &addVersionResp, httpclient.RequestOptions{}); err != nil {
		return nil, err
	}
	if !addVersionResp.IsSuccess() {
		return nil, fmt.Errorf("add release version error: %s", addVersionResp.Msg)
	}
	if r.Params.CodeUrl != "" {
		if err := r.clone(); err != nil {
			return nil, err
		}
	}
	if r.Params.Images != "" {
		if err := r.tagImage(); err != nil {
			return nil, err
		}
		imgBytes, _ := json.Marshal(r.Images)
		r.Result.Images = string(imgBytes)
	}
	return r.Result, nil
}

func (r *releaserExecutor) Cancel() error {
	r.canceled = true
	r.cancelFunc()
	return nil
}

func (r *releaserExecutor) clone() error {
	os.RemoveAll(r.CodeDir)
	r.Log("git clone %v", r.Params.CodeUrl)
	gitcli, err := utilgit.NewClient(r.Params.CodeType, r.Params.CodeApiUrl, r.Params.CodeSecret)
	if err != nil {
		return err
	}
	auth, err := gitcli.Auth()
	if err != nil {
		return err
	}
	repo, err := gitcli.Clone(r.ctx, r.CodeDir, false, &git.CloneOptions{
		URL:      r.Params.CodeUrl,
		Progress: r.Logger,
	})
	if err != nil {
		r.Log("克隆代码仓库失败：%v", err)
		klog.Errorf("job=%d clone %s error: %v", r.Params.JobId, r.Params.CodeUrl, err)
		return fmt.Errorf("git clone %s error: %v", r.Params.CodeUrl, err)
	}
	w, err := repo.Worktree()
	if err != nil {
		r.Log("克隆代码仓库失败：%v", err)
		klog.Errorf("job=%d clone %s error: %v", r.Params.JobId, r.Params.CodeUrl, err)
		return fmt.Errorf("git clone %s error: %v", r.Params.CodeUrl, err)
	}
	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(r.Params.CodeCommitId),
	})
	if err != nil {
		r.Log("git checkout %s 失败：%v", r.Params.CodeCommitId, err)
		klog.Errorf("job=%d git checkout %s error: %v", r.Params.JobId, r.Params.CodeCommitId, err)
		return fmt.Errorf("git checkout %s error: %v", r.Params.CodeCommitId, err)
	}
	r.Log("git tag %s", r.Params.Version)
	_, err = repo.CreateTag(r.Params.Version, plumbing.NewHash(r.Params.CodeCommitId), &git.CreateTagOptions{
		Message: r.Params.Version,
		Tagger: &object.Signature{
			Name:  "kubespace",
			Email: "kubespace@kubespace.cn",
			When:  time.Now(),
		},
	})
	if err != nil {
		r.Log("git tag error: %s", err.Error())
		if !errors.Is(err, git.ErrTagExists) {
			return fmt.Errorf("git tag error: %s", err.Error())
		}
	}
	po := &git.PushOptions{
		RemoteName: "origin",
		Progress:   r.Logger,
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/tags/*:refs/tags/*")},
		Auth:       auth,
	}
	r.Log("git push --tags")
	err = repo.PushContext(r.ctx, po)
	if err != nil {
		r.Log("git push error: %s", err.Error())
		return err
	}

	return nil
}

func (r *releaserExecutor) tagImage() error {
	r.Log("images=%s", r.Params.Images)
	images := stringToImage(r.Params.Images)
	r.Log("images=%v", images)
	if r.Params.ImageBuildRegistry.User != "" && r.Params.ImageBuildRegistry.Password != "" {
		if err := r.loginDocker(r.Params.ImageBuildRegistry.User, r.Params.ImageBuildRegistry.Password, r.Params.ImageBuildRegistry.Registry); err != nil {
			r.Log("docker login %s error: %v", r.Params.ImageBuildRegistry.Registry, err)
			klog.Errorf("docker login %s error: %v", r.Params.ImageBuildRegistry.Registry, err)
		}
	}
	for _, image := range images {
		if err := r.tagAndPushImage(image); err != nil {
			return err
		}
	}
	return nil
}

func (r *releaserExecutor) loginDocker(user string, password string, server string) error {
	r.Log("docker login %s", server)
	cmd := exec.CommandContext(r.ctx, "bash", "-c", fmt.Sprintf("docker login -u %s -p %s %s", user, password, server))
	cmd.Stdout = r.Logger
	cmd.Stderr = r.Logger
	return cmd.Run()
}

func (r *releaserExecutor) tagAndPushImage(image string) error {
	dockerBuildCmd := fmt.Sprintf("docker pull %s", image)
	cmd := exec.CommandContext(r.ctx, "bash", "-xc", dockerBuildCmd)
	cmd.Stdout = r.Logger
	cmd.Stderr = r.Logger
	if err := cmd.Run(); err != nil {
		r.Log("拉取镜像%s错误：%v", image, err)
		klog.Errorf("pull image error: %v", err)
		return fmt.Errorf("拉取镜像%s错误：%v", image, err)
	}
	newImage := strings.Split(image, ":")[0] + ":" + r.Params.Version
	cmd = exec.Command("bash", "-xc", "docker tag "+image+" "+newImage)
	cmd.Stdout = r.Logger
	cmd.Stderr = r.Logger
	if err := cmd.Run(); err != nil {
		r.Log("镜像打标签%s错误：%v", image, err)
		klog.Errorf("tag image error: %v", err)
		return fmt.Errorf("镜像打标签%s错误：%v", image, err)
	}
	if err := r.pushImage(newImage); err != nil {
		return err
	}
	r.Images[utils.GetImageName(newImage)] = newImage
	rmiImage := fmt.Sprintf("docker rmi %s && docker rmi %s", image, newImage)
	cmd = exec.CommandContext(r.ctx, "bash", "-xc", rmiImage)
	cmd.Stdout = r.Logger
	cmd.Stderr = r.Logger
	if err := cmd.Run(); err != nil {
		r.Log("删除本地镜像%s错误：%v", image, err)
		klog.Errorf("rmi image error: %v", err)
		return fmt.Errorf("删除本地构建镜像%s错误：%v", image, err)
	}
	return nil
}

func (r *releaserExecutor) pushImage(imageUrl string) error {
	pushCmd := fmt.Sprintf("docker push %s", imageUrl)
	cmd := exec.CommandContext(r.ctx, "bash", "-xc", pushCmd)
	cmd.Stdout = r.Logger
	cmd.Stderr = r.Logger
	if err := cmd.Run(); err != nil {
		r.Log("docker push %s：%v", imageUrl, err)
		klog.Errorf("push image error: %v", err)
		return fmt.Errorf("推送镜像%s错误：%v", imageUrl, err)
	}
	return nil
}
