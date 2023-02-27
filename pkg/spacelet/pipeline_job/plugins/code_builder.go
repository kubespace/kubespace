package plugins

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	sshgit "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/kubespace/kubespace/pkg/utils"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	CodeBuildTypeNone   = "none"
	CodeBuildTypeFile   = "file"
	CodeBuildTypeScript = "script"
)

type CodeBuilderPlugin struct{}

func (b CodeBuilderPlugin) Execute(params *PluginParams) (interface{}, error) {
	buildCodePlugin, err := newCodeBuilderPlugin(params)
	if err != nil {
		return nil, err
	}

	return buildCodePlugin.execute()
}

type Secret struct {
	Type        string `json:"type"`
	User        string `json:"user"`
	Password    string `json:"password"`
	PrivateKey  string `json:"private_key"`
	AccessToken string `json:"access_token"`
}

type PipelineResource struct {
	Type   string `json:"type"`
	Value  string `json:"value"`
	Secret Secret `json:"secret"`
}

type ImageBuilds struct {
	Dockerfile string `json:"dockerfile"`
	Image      string `json:"image"`
}

type ImageRegistry struct {
	Registry string `json:"registry"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type codeBuilderParams struct {
	JobId uint `json:"job_id"`

	CodeUrl         string           `json:"code_url"`
	CodeBranch      string           `json:"code_branch"`
	CodeCommitId    string           `json:"code_commit_id"`
	CodeSecret      Secret           `json:"code_secret"`
	CodeBuild       bool             `json:"code_build"`
	CodeBuildType   string           `json:"code_build_type"`
	CodeBuildImage  PipelineResource `json:"code_build_image"`
	CodeBuildFile   string           `json:"code_build_file"`
	CodeBuildScript string           `json:"code_build_script"`
	CodeBuildExec   string           `json:"code_build_exec"`

	ImageBuildRegistryId int           `json:"image_registry_id"`
	ImageBuildRegistry   ImageRegistry `json:"image_build_registry"`
	ImageBuilds          []ImageBuilds `json:"image_builds"`
}

type codeBuilderPlugin struct {
	*JobLogger
	Params  *codeBuilderParams
	DataDir string
	CodeDir string
	Images  []string
	Result  *CodeBuilderPluginResult
}

type CodeBuilderPluginResult struct {
	ImageUrl        string `json:"images"`
	ImageRegistry   string `json:"image_registry"`
	ImageRegistryId int    `json:"image_registry_id"`
}

func newCodeBuilderPlugin(params *PluginParams) (*codeBuilderPlugin, error) {
	var buildParams codeBuilderParams
	if err := utils.ConvertTypeByJson(params.Params, &buildParams); err != nil {
		return nil, err
	}
	if buildParams.ImageBuildRegistry.Registry == "" {
		buildParams.ImageBuildRegistry.Registry = "docker.io"
	}
	buildCodePlugin := &codeBuilderPlugin{
		Params:    &buildParams,
		JobLogger: params.Logger,
		Result: &CodeBuilderPluginResult{
			ImageUrl:        "",
			ImageRegistryId: buildParams.ImageBuildRegistryId,
			ImageRegistry:   buildParams.ImageBuildRegistry.Registry,
		},
	}
	codeDir := utils.GetCodeRepoName(buildParams.CodeUrl)
	if codeDir == "" {
		klog.Errorf("job=%d get empty code repo name", buildParams.JobId)
		return nil, fmt.Errorf("get empty code repo name")
	}
	buildCodePlugin.CodeDir, _ = filepath.Abs(filepath.Join(params.RootDir, codeDir))

	return buildCodePlugin, nil
}

func (b *codeBuilderPlugin) execute() (interface{}, error) {
	if err := b.clone(); err != nil {
		return nil, err
	}
	if err := b.buildCode(); err != nil {
		return nil, err
	}
	if err := b.buildImages(); err != nil {
		return nil, err
	}
	return b.Result, nil
}

func (b *codeBuilderPlugin) clone() error {
	if err := os.RemoveAll(b.CodeDir); err != nil {
		return err
	}
	b.Log("git clone %v", b.Params.CodeUrl)
	time.Sleep(1)
	var auth transport.AuthMethod
	var err error
	if b.Params.CodeSecret.Type == "key" {
		privateKey, err := sshgit.NewPublicKeys("git", []byte(b.Params.CodeSecret.PrivateKey), "")
		if err != nil {
			return fmt.Errorf("生成代码密钥失败：" + err.Error())
		}
		privateKey.HostKeyCallbackHelper = sshgit.HostKeyCallbackHelper{
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		auth = privateKey
	} else if b.Params.CodeSecret.Type == "password" {
		auth = &http.BasicAuth{
			Username: b.Params.CodeSecret.User,
			Password: b.Params.CodeSecret.Password,
		}
	}
	r, err := git.PlainClone(b.CodeDir, false, &git.CloneOptions{
		Auth:     auth,
		URL:      b.Params.CodeUrl,
		Progress: b.JobLogger,
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

func (b *codeBuilderPlugin) buildCode() error {
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
		if err := os.WriteFile(filepath.Join(b.CodeDir, ".build.sh"), []byte(b.Params.CodeBuildScript), 0666); err != nil {
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
		shExec = "bash"
	}

	dockerRunCmd := fmt.Sprintf("docker run --net=host --rm -i -v %s:/app -w /app --entrypoint sh %s -c \"%s -ex /app/%s 2>&1\"", b.CodeDir, b.Params.CodeBuildImage.Value, shExec, codeBuildFile)
	klog.Infof("job=%d code build cmd: %s", b.Params.JobId, dockerRunCmd)
	cmd := exec.Command("bash", "-xc", dockerRunCmd)
	cmd.Stdout = b.JobLogger
	cmd.Stderr = b.JobLogger
	if err := cmd.Run(); err != nil {
		klog.Errorf("job=%d build error: %v", b.Params.JobId, err)
		return fmt.Errorf("build code error: %v", err)
	}
	return nil
}

func (b *codeBuilderPlugin) buildImages() error {
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
	b.Result.ImageUrl = strings.Join(b.Images, ",")
	return nil
}

func (b *codeBuilderPlugin) buildAndPushImage(dockerfilePath string, imageName string) error {
	dockerfile := filepath.Join(b.CodeDir, dockerfilePath)
	//baseDockerfile := filepath.Dir(dockerfile)
	dockerBuildCmd := fmt.Sprintf("docker build -t %s -f %s %s", imageName, dockerfile, b.CodeDir)
	cmd := exec.Command("bash", "-xc", dockerBuildCmd)
	cmd.Stdout = b.JobLogger
	cmd.Stderr = b.JobLogger
	if err := cmd.Run(); err != nil {
		b.Log("构建镜像%s错误：%v", imageName, err)
		klog.Errorf("build image error: %v", err)
		return fmt.Errorf("构建镜像%s错误：%v", imageName, err)
	}
	if err := b.pushImage(imageName); err != nil {
		return err
	}
	b.Images = append(b.Images, imageName)
	cmd = exec.Command("bash", "-xc", "docker rmi "+imageName)
	cmd.Stdout = b.JobLogger
	cmd.Stderr = b.JobLogger
	if err := cmd.Run(); err != nil {
		b.Log("删除本地镜像%s错误：%v", imageName, err)
		klog.Errorf("remove image %s error: %v", imageName, err)
		//return fmt.Errorf("删除本地构建镜像%s错误：%v", imageName, err)
	}
	return nil
}

func (b *codeBuilderPlugin) loginDocker(user string, password string, server string) error {
	b.Log("docker login %s", server)
	cmd := exec.Command("bash", "-c", fmt.Sprintf("docker login -u %s -p %s %s", user, password, server))
	cmd.Stdout = b.JobLogger
	cmd.Stderr = b.JobLogger
	return cmd.Run()
}

func (b *codeBuilderPlugin) pushImage(imageUrl string) error {
	if b.Params.ImageBuildRegistry.User != "" && b.Params.ImageBuildRegistry.Password != "" {
		if err := b.loginDocker(b.Params.ImageBuildRegistry.User, b.Params.ImageBuildRegistry.Password, b.Params.ImageBuildRegistry.Registry); err != nil {
			b.Log("docker login %s error: %v", b.Params.ImageBuildRegistry.Registry, err)
			klog.Errorf("docker login %s error: %v", b.Params.ImageBuildRegistry.Registry, err)
		}
	}
	pushCmd := fmt.Sprintf("docker push %s", imageUrl)
	cmd := exec.Command("bash", "-xc", pushCmd)
	cmd.Stdout = b.JobLogger
	cmd.Stderr = b.JobLogger
	if err := cmd.Run(); err != nil {
		b.Log("docker push %s：%v", imageUrl, err)
		klog.Errorf("push image error: %v", err)
		return fmt.Errorf("推送镜像%s错误：%v", imageUrl, err)
	}
	return nil
}
