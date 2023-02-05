package plugins

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	sshgit "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/utils"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ReleaserPlugin struct {
	Models *model.Models
}

func (b ReleaserPlugin) Execute(params *PluginParams) (interface{}, error) {
	releasePlugin, err := newReleaserPlugin(params, b.Models)
	if err != nil {
		return nil, err
	}

	return releasePlugin.execute()
}

type releaseParams struct {
	JobId       uint `json:"job_id"`
	WorkspaceId uint `json:"workspace_id"`

	CodeUrl      string  `json:"code_url"`
	CodeBranch   string  `json:"code_branch"`
	CodeCommitId string  `json:"code_commit_id"`
	CodeSecret   *Secret `json:"code_secret"`

	ImageBuildRegistry ImageRegistry `json:"image_registry"`

	Version string `json:"version"`
	Images  string `json:"images"`
}

type ReleaserPluginResult struct {
	Version string `json:"version"`
	Images  string `json:"images"`
}

type releaserPlugin struct {
	*PluginLogger
	models  *model.Models
	Params  *releaseParams
	CodeDir string
	Result  *ReleaserPluginResult
	Images  []string
	rootDir string
}

func newReleaserPlugin(pluginParams *PluginParams, models *model.Models) (*releaserPlugin, error) {
	var params releaseParams
	if err := utils.ConvertTypeByJson(pluginParams.Params, &params); err != nil {
		return nil, err
	}
	if params.Version == "" {
		return nil, fmt.Errorf("发布版本号为空")
	}
	plugin := &releaserPlugin{
		models:       models,
		PluginLogger: pluginParams.Logger,
		Params:       &params,
		Result: &ReleaserPluginResult{
			Version: params.Version,
			Images:  "",
		},
	}
	codeDir := utils.GetCodeRepoName(params.CodeUrl)
	if codeDir == "" {
		klog.Errorf("job=%d get empty code repo name", params.JobId)
		return nil, fmt.Errorf("get empty code repo name")
	}
	rootDir := filepath.Join(pluginParams.DataDir, "pipeline", strconv.Itoa(int(params.JobId)))
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return nil, err
	}
	plugin.CodeDir, _ = filepath.Abs(filepath.Join(rootDir, codeDir))
	plugin.rootDir = rootDir

	return plugin, nil
}

func (r *releaserPlugin) execute() (interface{}, error) {
	defer func() {
		if err := os.RemoveAll(r.rootDir); err != nil {
			r.Log("remove job root dir %s error: %s", r.rootDir, err.Error())
		}
	}()
	err := r.models.PipelineReleaseManager.Add(r.Params.WorkspaceId, r.Params.Version, r.Params.JobId)
	if err != nil {
		return nil, err
	}
	if r.Params.CodeUrl != "" {
		if err = r.clone(); err != nil {
			return nil, err
		}
	}
	if r.Params.Images != "" {
		if err = r.tagImage(); err != nil {
			return nil, err
		}
		r.Result.Images = strings.Join(r.Images, ",")
	}
	return r.Result, nil
}

func (r *releaserPlugin) clone() error {
	os.RemoveAll(r.CodeDir)
	r.Log("git clone %v", r.Params.CodeUrl)
	time.Sleep(1)
	var auth transport.AuthMethod
	var err error
	if r.Params.CodeSecret.Type == "key" {
		privateKey, err := sshgit.NewPublicKeys("git", []byte(r.Params.CodeSecret.PrivateKey), "")
		if err != nil {
			return fmt.Errorf("生成代码密钥失败：" + err.Error())
		}
		privateKey.HostKeyCallbackHelper = sshgit.HostKeyCallbackHelper{
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		auth = privateKey
	} else if r.Params.CodeSecret.Type == "password" {
		auth = &http.BasicAuth{
			Username: r.Params.CodeSecret.User,
			Password: r.Params.CodeSecret.Password,
		}
	}
	repo, err := git.PlainClone(r.CodeDir, false, &git.CloneOptions{
		Auth:     auth,
		URL:      r.Params.CodeUrl,
		Progress: r.PluginLogger,
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
		Progress:   r.PluginLogger,
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/tags/*:refs/tags/*")},
		Auth:       auth,
	}
	r.Log("git push --tags")
	err = repo.Push(po)
	if err != nil {
		r.Log("git push error: %s", err.Error())
		return err
	}

	return nil
}

func (r *releaserPlugin) tagImage() error {
	images := strings.Split(r.Params.Images, ",")
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

func (r *releaserPlugin) loginDocker(user string, password string, server string) error {
	r.Log("docker login %s", server)
	cmd := exec.Command("bash", "-c", fmt.Sprintf("docker login -u %s -p %s %s", user, password, server))
	cmd.Stdout = r.PluginLogger
	cmd.Stderr = r.PluginLogger
	return cmd.Run()
}

func (r *releaserPlugin) tagAndPushImage(image string) error {
	dockerBuildCmd := fmt.Sprintf("docker pull %s", image)
	cmd := exec.Command("bash", "-xc", dockerBuildCmd)
	cmd.Stdout = r.PluginLogger
	cmd.Stderr = r.PluginLogger
	if err := cmd.Run(); err != nil {
		r.Log("拉取镜像%s错误：%v", image, err)
		klog.Errorf("pull image error: %v", err)
		return fmt.Errorf("拉取镜像%s错误：%v", image, err)
	}
	newImage := strings.Split(image, ":")[0] + ":" + r.Params.Version
	cmd = exec.Command("bash", "-xc", "docker tag "+image+" "+newImage)
	cmd.Stdout = r.PluginLogger
	cmd.Stderr = r.PluginLogger
	if err := cmd.Run(); err != nil {
		r.Log("镜像打标签%s错误：%v", image, err)
		klog.Errorf("tag image error: %v", err)
		return fmt.Errorf("镜像打标签%s错误：%v", image, err)
	}
	if err := r.pushImage(newImage); err != nil {
		return err
	}
	r.Images = append(r.Images, newImage)
	rmiImage := fmt.Sprintf("docker rmi %s && docker rmi %s", image, newImage)
	cmd = exec.Command("bash", "-xc", rmiImage)
	cmd.Stdout = r.PluginLogger
	cmd.Stderr = r.PluginLogger
	if err := cmd.Run(); err != nil {
		r.Log("删除本地镜像%s错误：%v", image, err)
		klog.Errorf("rmi image error: %v", err)
		return fmt.Errorf("删除本地构建镜像%s错误：%v", image, err)
	}
	return nil
}

func (r *releaserPlugin) pushImage(imageUrl string) error {
	pushCmd := fmt.Sprintf("docker push %s", imageUrl)
	cmd := exec.Command("bash", "-xc", pushCmd)
	cmd.Stdout = r.PluginLogger
	cmd.Stderr = r.PluginLogger
	if err := cmd.Run(); err != nil {
		r.Log("docker push %s：%v", imageUrl, err)
		klog.Errorf("push image error: %v", err)
		return fmt.Errorf("推送镜像%s错误：%v", imageUrl, err)
	}
	return nil
}
