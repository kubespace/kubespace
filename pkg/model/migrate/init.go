package migrate

// 每次新加迁移版本，需要在这里初始化，注册到迁移列表
import (
	_ "github.com/kubespace/kubespace/pkg/model/migrate/v1_1/v1_1_2_a_pipeline_workspace_chg_code"
	"github.com/kubespace/kubespace/pkg/model/types"
)

var initTypes = []interface{}{
	&types.Cluster{},
	&types.User{},
	&types.UserRole{},
	&types.PipelineWorkspace{},
	&types.Pipeline{},
	&types.PipelineStage{},
	&types.PipelinePlugin{},
	&types.PipelineRun{},
	&types.PipelineRunStage{},
	&types.PipelineRunJob{},
	&types.PipelineRunJobLog{},
	&types.PipelineResource{},
	&types.PipelineWorkspaceRelease{},

	&types.SettingsSecret{},
	&types.SettingsImageRegistry{},

	&types.Project{},
	&types.ProjectApp{},
	&types.AppVersion{},
	&types.AppVersionChart{},
	&types.AppStore{},
	&types.ProjectAppRevision{},
	&types.Ldap{},
}
