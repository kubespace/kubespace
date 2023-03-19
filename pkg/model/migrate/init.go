package migrate

import (
	_ "github.com/kubespace/kubespace/pkg/model/migrate/v1_1/v1_1_2_a_pipeline_workspace_chg_code"
	_ "github.com/kubespace/kubespace/pkg/model/migrate/v1_1/v1_1_3_a_add_spacelet"
	_ "github.com/kubespace/kubespace/pkg/model/migrate/v1_1/v1_1_3_b_add_ldap"
	_ "github.com/kubespace/kubespace/pkg/model/migrate/v1_1/v1_1_4_a_stagerun_add_finishtime"
	_ "github.com/kubespace/kubespace/pkg/model/migrate/v1_1/v1_1_5_a_chg_pipeline_source_column"
	"github.com/kubespace/kubespace/pkg/model/types"
)

// 注意：每次新加迁移版本，需要在这里初始化，注册到迁移列表
// ！！！ 同时在import里添加migrate ！！！
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
	&types.Spacelet{},
	&types.Ldap{},
}
