package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/kubespace/kubespace/pkg/model/types"
	utilgit "github.com/kubespace/kubespace/pkg/third/git"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// 代码构建镜像生成的镜像列表从字符串转换为map对象
// 如{"kubespace/kubespace": "docker.io/kubespace/kubespace:latest"}
func stringToImage(imgStr string) map[string]string {
	if imgStr == "" {
		return nil
	}
	imgMap := make(map[string]string)
	imgSplits := strings.Split(imgStr, ",{")
	for _, imgSplitStr := range imgSplits {
		if imgSplitStr[0:1] != "{" {
			imgSplitStr = "{" + imgSplitStr
		}
		imgSplitMap := make(map[string]string)
		if err := json.Unmarshal([]byte(imgSplitStr), &imgSplitMap); err != nil {
			continue
		}
		for k, v := range imgSplitMap {
			imgMap[k] = v
		}
	}
	return imgMap
}

const (
	CodeBuildTypeNone   = "none"
	CodeBuildTypeFile   = "file"
	CodeBuildTypeScript = "script"
)

// CodeBuilderPlugin 流水线代码构建执行器
type CodeBuilderPlugin struct{}

func (b CodeBuilderPlugin) Executor(params *ExecutorParams) (Executor, error) {
	return newCodeBuilderExecutor(params)
}

type PipelineResource struct {
	Type   string        `json:"type"`
	Value  string        `json:"value"`
	Secret *types.Secret `json:"secret"`
}

type ImageBuilds struct {
	Dockerfile string `json:"dockerfile"`
	Image      string `json:"image"`
}

type codeBuilderParams struct {
	JobId uint `json:"job_id"`

	CodeUrl         string           `json:"code_url"`
	CodeApiUrl      string           `json:"code_api_url"`
	CodeType        string           `json:"code_type"`
	CodeBranch      string           `json:"code_branch"`
	CodeCommitId    string           `json:"code_commit_id"`
	CodeSecret      *types.Secret    `json:"code_secret"`
	CodeBuild       bool             `json:"code_build"`
	CodeBuildType   string           `json:"code_build_type"`
	CodeBuildImage  PipelineResource `json:"code_build_image"`
	CodeBuildFile   string           `json:"code_build_file"`
	CodeBuildScript string           `json:"code_build_script"`
	CodeBuildExec   string           `json:"code_build_exec"`

	ImageBuildRegistryId int                 `json:"image_registry_id"`
	ImageBuildRegistry   types.ImageRegistry `json:"image_build_registry"`
	ImageBuilds          []ImageBuilds       `json:"image_builds"`
}

// CodeBuilderPluginResult 代码构建执行结果
type CodeBuilderPluginResult struct {
	ImageUrl        string `json:"images"`
	ImageRegistry   string `json:"image_registry"`
	ImageRegistryId int    `json:"image_registry_id"`
}

type codeBuilderExecutor struct {
	Logger
	Params  *codeBuilderParams
	codeDir string
	images  map[string]string
	result  *CodeBuilderPluginResult
	// 是否已取消
	canceled   bool
	cancelFunc context.CancelFunc
	ctx        context.Context
}

func newCodeBuilderExecutor(params *ExecutorParams) (*codeBuilderExecutor, error) {
	var buildParams codeBuilderParams
	if err := utils.ConvertTypeByJson(params.Params, &buildParams); err != nil {
		return nil, err
	}
	if buildParams.ImageBuildRegistry.Registry == "" {
		buildParams.ImageBuildRegistry.Registry = "docker.io"
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	buildCodePlugin := &codeBuilderExecutor{
		ctx:        ctx,
		cancelFunc: cancelFunc,
		Params:     &buildParams,
		Logger:     params.Logger,
		images:     make(map[string]string),
		result: &CodeBuilderPluginResult{
			ImageRegistryId: buildParams.ImageBuildRegistryId,
			ImageRegistry:   buildParams.ImageBuildRegistry.Registry,
		},
	}
	codeDir := utils.GetCodeRepoName(buildParams.CodeUrl)
	if codeDir == "" {
		klog.Errorf("job=%d get empty code repo name", buildParams.JobId)
		return nil, fmt.Errorf("get empty code repo name")
	}
	buildCodePlugin.codeDir, _ = filepath.Abs(filepath.Join(params.RootDir, codeDir))

	return buildCodePlugin, nil
}

func (b *codeBuilderExecutor) Execute() (interface{}, error) {
	steps := []stepFunc{b.clone, b.buildCode, b.buildImages}
	for _, step := range steps {
		err := step()
		if b.canceled {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
	}
	return b.result, nil
}

func (b *codeBuilderExecutor) Cancel() error {
	b.canceled = true
	b.cancelFunc()
	return nil
}

// 克隆代码
func (b *codeBuilderExecutor) clone() error {
	if err := os.RemoveAll(b.codeDir); err != nil {
		return err
	}
	b.Log("git clone %v", b.Params.CodeUrl)
	time.Sleep(1)
	var err error

	gitcli, err := utilgit.NewClient(b.Params.CodeType, b.Params.CodeApiUrl, b.Params.CodeSecret)
	if err != nil {
		return err
	}
	r, err := gitcli.Clone(b.ctx, b.codeDir, false, &git.CloneOptions{
		URL:      b.Params.CodeUrl,
		Progress: b.Logger,
	})
	if err != nil {
		b.Log("克隆代码仓库失败：%v", err)
		klog.Errorf("job=%d clone %s error: %v", b.Params.JobId, b.Params.CodeUrl, err)
		return fmt.Errorf("git clone %s error: %v", b.Params.CodeUrl, err)
	}
	w, err := r.Worktree()
	if err != nil {
		b.Log("克隆代码仓库失败：%v", err)
		klog.Errorf("job=%d clone %s error: %v", b.Params.JobId, b.Params.CodeUrl, err)
		return fmt.Errorf("git clone %s error: %v", b.Params.CodeUrl, err)
	}
	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(b.Params.CodeCommitId),
	})
	if err != nil {
		b.Log("git checkout %s 失败：%v", b.Params.CodeCommitId, err)
		klog.Errorf("job=%d git checkout %s error: %v", b.Params.JobId, b.Params.CodeCommitId, err)
		return fmt.Errorf("git checkout %s error: %v", b.Params.CodeCommitId, err)
	}
	return nil
}

// 代码编译
func (b *codeBuilderExecutor) buildCode() error {
	if b.Params.CodeBuildType == CodeBuildTypeNone || !b.Params.CodeBuild {
		b.Log("跳过代码构建")
		return nil
	}
	if b.Params.CodeBuildImage.Value == "" {
		b.Log("构建代码镜像为空，请检查流水线配置")
		return fmt.Errorf("build code image is empty")
	}
	codeBuildFile := ".build.sh"
	codeBuildType := b.Params.CodeBuildType
	if codeBuildType == "" {
		codeBuildType = CodeBuildTypeScript
	}
	if b.Params.CodeBuildType == CodeBuildTypeScript {
		if b.Params.CodeBuildScript == "" {
			b.Log("代码构建脚本为空")
			return nil
		}
		if err := os.WriteFile(filepath.Join(b.codeDir, ".build.sh"), []byte(b.Params.CodeBuildScript), 0666); err != nil {
			b.Log("写脚本文件%s错误：%v", codeBuildFile, err)
			klog.Errorf("job=%d write build error: %v", b.Params.JobId, err)
			return fmt.Errorf("write build file error: %s", err.Error())
		}
	} else if b.Params.CodeBuildType == CodeBuildTypeFile {
		codeBuildFile = b.Params.CodeBuildFile
	}
	if codeBuildFile == "" {
		codeBuildFile = "build.sh"
	}
	shExec := b.Params.CodeBuildExec
	if shExec == "" {
		shExec = "sh"
	}

	dockerRunCmd := fmt.Sprintf("docker run --net=host --rm -i -v %s:/app -w /app --entrypoint sh %s -c \"%s -ex /app/%s 2>&1\"", b.codeDir, b.Params.CodeBuildImage.Value, shExec, codeBuildFile)
	klog.Infof("job=%d code build cmd: %s", b.Params.JobId, dockerRunCmd)
	cmd := exec.CommandContext(b.ctx, "bash", "-xc", dockerRunCmd)
	cmd.Stdout = b.Logger
	cmd.Stderr = b.Logger
	if err := cmd.Run(); err != nil {
		klog.Errorf("job=%d build error: %v", b.Params.JobId, err)
		b.Log("build error: %s", err.Error())
		return fmt.Errorf("build code error: %v", err)
	}
	return nil
}

// 构建镜像
func (b *codeBuilderExecutor) buildImages() error {
	timeStr := fmt.Sprintf("%d", time.Now().Unix())
	for _, buildImage := range b.Params.ImageBuilds {
		imageName := buildImage.Image
		if imageName == "" {
			b.Log("not found build image parameter")
			return fmt.Errorf("not found build image parameter")
		}
		imageName = strings.Split(imageName, ":")[0]
		if b.Params.ImageBuildRegistry.Registry != "" {
			imageName = b.Params.ImageBuildRegistry.Registry + "/" + imageName + ":" + timeStr
		} else {
			imageName = "docker.io/" + imageName + ":" + timeStr
		}
		dockerfile := buildImage.Dockerfile
		if dockerfile == "" {
			dockerfile = "Dockerfile"
		}
		if err := b.buildAndPushImage(dockerfile, imageName); err != nil {
			return err
		}
	}
	imgs, _ := json.Marshal(b.images)
	b.result.ImageUrl = string(imgs)
	return nil
}

func (b *codeBuilderExecutor) buildAndPushImage(dockerfilePath string, imageName string) error {
	dockerfile := filepath.Join(b.codeDir, dockerfilePath)
	//baseDockerfile := filepath.Dir(dockerfile)
	dockerBuildCmd := fmt.Sprintf("docker build -t %s -f %s %s", imageName, dockerfile, b.codeDir)
	cmd := exec.CommandContext(b.ctx, "bash", "-xc", dockerBuildCmd)
	cmd.Stdout = b.Logger
	cmd.Stderr = b.Logger
	if err := cmd.Run(); err != nil {
		b.Log("构建镜像%s错误：%v", imageName, err)
		klog.Errorf("build image error: %v", err)
		return fmt.Errorf("构建镜像%s错误：%v", imageName, err)
	}
	if err := b.pushImage(imageName); err != nil {
		return err
	}
	b.images[utils.GetImageName(imageName)] = imageName
	cmd = exec.CommandContext(b.ctx, "bash", "-xc", "docker rmi "+imageName)
	cmd.Stdout = b.Logger
	cmd.Stderr = b.Logger
	if err := cmd.Run(); err != nil {
		b.Log("删除本地镜像%s错误：%v", imageName, err)
		klog.Errorf("remove image %s error: %v", imageName, err)
		//return fmt.Errorf("删除本地构建镜像%s错误：%v", imageName, err)
	}
	return nil
}

func (b *codeBuilderExecutor) loginDocker(user string, password string, server string) error {
	b.Log("docker login %s", server)
	cmd := exec.CommandContext(b.ctx, "bash", "-c", fmt.Sprintf("docker login -u %s -p %s %s", user, password, server))
	cmd.Stdout = b.Logger
	cmd.Stderr = b.Logger
	return cmd.Run()
}

func (b *codeBuilderExecutor) pushImage(imageUrl string) error {
	if b.Params.ImageBuildRegistry.User != "" && b.Params.ImageBuildRegistry.Password != "" {
		if err := b.loginDocker(b.Params.ImageBuildRegistry.User, b.Params.ImageBuildRegistry.Password, b.Params.ImageBuildRegistry.Registry); err != nil {
			b.Log("docker login %s error: %v", b.Params.ImageBuildRegistry.Registry, err)
			klog.Errorf("docker login %s error: %v", b.Params.ImageBuildRegistry.Registry, err)
		}
	}
	pushCmd := fmt.Sprintf("docker push %s", imageUrl)
	cmd := exec.CommandContext(b.ctx, "bash", "-xc", pushCmd)
	cmd.Stdout = b.Logger
	cmd.Stderr = b.Logger
	if err := cmd.Run(); err != nil {
		b.Log("docker push %s：%v", imageUrl, err)
		klog.Errorf("push image error: %v", err)
		return fmt.Errorf("推送镜像%s错误：%v", imageUrl, err)
	}
	return nil
}
