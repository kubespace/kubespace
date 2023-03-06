package main

import (
	"context"
	"github.com/kubespace/kubespace/pkg/model/types"
	utilgit "github.com/kubespace/kubespace/pkg/utils/git"
	"k8s.io/klog/v2"
)

func main() {
	gitcli, err := utilgit.NewClient(types.WorkspaceCodeTypeGitLab, "https://gitlab.com",
		&utilgit.Secret{AccessToken: "glpat-sx3K4i1iEc_i54QuZYZu"})
	if err != nil {
		klog.Fatal(err)
	}
	repos, err := gitcli.ListRepositories(context.Background())
	if err != nil {
		klog.Fatal(err)
	}
	klog.Info(repos)
	//err = gitcli.CreateTag(
	//	context.Background(),
	//	"https://gitlab.com/kubespace/kubespace.git",
	//	"190a4229e620e0f434922473a917cdd0ea3b41bf",
	//	"v1.1.1")
	//klog.Info(err)
	//
	//gitcli.Clone(context.Background(), "/Users/tomlee/data/abc", false, &git.CloneOptions{
	//	URL:             "https://gitlab.com/kubespace/kubespace.git",
	//	Progress:        os.Stdout,
	//	InsecureSkipTLS: true,
	//})
	branches, _ := gitcli.ListRepoBranches(context.Background(), "https://gitlab.com/kubespace/kubespace.git")
	klog.Info(branches)
	commit, _ := gitcli.GetBranchLatestCommit(context.Background(), "https://gitlab.com/kubespace/kubespace.git", "main")
	klog.Info(commit)
}
