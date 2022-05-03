<template>
  <div>
    <clusterbar :titleName="titleName" :titleLink="['pipeline']">
      <div slot="right-btn">
        <el-button size="small" class="bar-btn" type="" @click="cancelEdit">取消</el-button>
        <el-button size="small" class="bar-btn" type="primary" @click="savePipeline">保存</el-button>
      </div>
    </clusterbar>
    <div v-loading="loading" class="dashboard-container" :style="{height: maxHeight + 'px'}" :max-height="maxHeight">
      <div style="padding: 10px 8px 0px;">
        <div>基本信息</div>
        <el-form label-position="left" class="pipeline-form-item" label-width="80px">
          <el-form-item label="名称">
            <el-input v-model="editPipeline.name" autocomplete="off" placeholder="请输入流水线名称" size="small"></el-input>
          </el-form-item>
        </el-form>
      </div>
      <div style="padding: 10px 8px 0px;" >
        <div>阶段任务</div>
        <div class="stage-job-outer">
          <div class="stage-job-line">
            <div class="pipeline-source-outer" style="display:inline-block" @click="openEditSource()">
              <div v-if="workspace.type == 'code'">
                <span class="pipeline-source-outer__span">
                  代码库源
                </span>
                <div style="font-size: 12px; padding: 10px 0px 0px; font-weight: 450">
                  {{ workspace ? workspace.code_url : '' }}
                </div>
                <div v-for="(t, i) in editPipeline.triggers" :key="i"
                  style="font-size: 12px; padding: 5px 10px 0px; font-weight: 400" >
                  <svg-icon icon-class="branch" /> {{ operatorMap[t.operator] }} {{ t.branch }}
                </div>
              </div>
              <div v-else>
                <span class="pipeline-source-outer__span">
                  流水线源
                </span>
                <template v-if="editPipeline.triggers && editPipeline.triggers.length > 0">
                  <div v-for="(t, i) in editPipeline.triggers" :key="i">
                    <div style="font-size: 12px; padding: 10px 0px 0px; font-weight: 450; width: 200px">
                      {{ t.workspace_name ? t.workspace_name : getTriggerWorkspaceName(t.workspace) }}
                    </div>
                    <div style="font-size: 12px; padding: 5px 20px 0px; font-weight: 400" >
                      — {{ t.pipeline_name ? t.pipeline_name : getTriggerPipelineName(t.workspace, t.pipeline) }}
                    </div>
                  </div>
                </template>
                <template v-else>
                  <div style="font-size: 12px; padding: 10px 0px 0px; font-weight: 450; width: 200px">
                      点击添加代码流水线源
                    </div>
                </template>
              </div>
            </div>
          </div>
          <div style="display: inline-block;">
            <div style="margin-top: 43px; width: 38px;">
              <div class="stage-job-line__inner" style="width: 20px;"></div>
              <div class="stage-job-line__add" @click="openAddStageDialog(0); dialogVisible=true;">
                <el-tooltip class="item" effect="light" content="添加阶段" placement="top" :hide-after="2000">
                  <i class="el-icon-circle-plus"></i>
                </el-tooltip>
              </div>
            </div>
          </div>
          <div v-for="(stage, i) in editPipeline.stages" :key="i">
            <div class="stage-job-line">
              <div class="stage-job-block">
                <div class="stage-job-block__stage" @click="openEditStageDialog(stage, i); dialogVisible=true;">
                  <span>#{{ i + 1 }} {{ stage.name }}</span>
                </div>
              </div>
              <div>
                <div class="stage-job-line__inner"></div>
                <a :class="checkJobError(stage.jobs[0]) ? 'stage-job-line__circle' : 'stage-job-line__circle-error'" @click="openEditJobDialog(stage, 0); dialogVisible=true;">
                  {{ checkJobError(stage.jobs[0]) ? '' : '!' }}
                </a>
                <div class="stage-job-line__inner"></div>
                <div class="stage-job-line__add" @click="openAddStageDialog(i+1); dialogVisible=true;">
                  <el-tooltip class="item" effect="light" content="添加阶段" placement="top" :hide-after="2000">
                    <i class="el-icon-circle-plus"></i>
                  </el-tooltip>
                </div>
              </div>
              <div class="stage-job-block">
                <div class="stage-job-block__job-name">{{ stage.jobs[0].name }}</div>
              </div>
              <div class="stage-job-block">
                <template  v-if="stage.jobs && stage.jobs.length > 1" >
                  <div v-for="(job, ji) in stage.jobs" :key="ji">
                    <template v-if="ji >0">
                      <div class="stage-job-block__job">
                        <div class="stage-job-block__job-circle">
                           <div :class="checkJobError(job) ? 'stage-job-line__circle' : 'stage-job-line__circle-error'" 
                            @click="openEditJobDialog(stage, ji); dialogVisible=true;">
                            <span style="margin-top: -5px;">{{ checkJobError(job) ? '' : '!' }}</span>
                          </div>
                        </div>
                      </div>
                      <div class="stage-job-block__job-name" style="margin-top: 18px;">{{ job.name }}</div>
                    </template>
                  </div>
                </template>

                <div class="stage-job-block__job-add">
                  <div class="stage-job-block__job-add__inner">
                    <span class="stage-job-block__job-add__inner-name" @click="openAddJobDialog(stage); dialogVisible=true;">+ 新建并行任务</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

        </div>
      </div>
    </div>

    <el-dialog :title="dialogTitleMap[dialogType]" :visible.sync="dialogVisible" :destroy-on-close="true" 
      @close="dialogType=''; dialogData={}" top="3vh" width="70%" :close-on-click-modal="false">
      <div class="dialogContent" style="padding: 0px 30px;">
        <template v-if="dialogType == 'edit_stage' || dialogType == 'add_stage'">
          <pipeline-stage :stage="dialogData"></pipeline-stage>
        </template>
        <template v-if="dialogType == 'edit_job' || dialogType == 'add_job'">
          <el-form :model="dialogData" ref="stage" label-position="left" label-width="105px">
            <el-form-item label="任务名称" prop="" :required="true">
              <el-input style="width: 250px;" v-model="dialogData.name" autocomplete="off" size="small"></el-input>
            </el-form-item>
            <el-form-item label="任务插件" prop="" :required="true">
              <el-select style="width: 250px;" v-model="dialogData.plugin_key" placeholder="任务插件" size="small"
                @change="dialogData.params={}">
                <el-option
                  v-for="plugin in jobPlugins"
                  :key="plugin.key"
                  :label="plugin.name"
                  :value="plugin.key">
                </el-option>
              </el-select>
            </el-form-item>
            <div v-if="jobPluginMap[dialogData.plugin_key]" style="background-color: #F2F6FC; padding: 15px;">
              <component v-if="jobPluginMap[dialogData.plugin_key]" v-bind:is="jobPluginMap[dialogData.plugin_key].component" :params="dialogData.params"></component>
            </div>
          </el-form>
        </template>
        <template v-if="dialogType == 'source'">
          <el-form :model="dialogData" label-position="left" label-width="105px" v-if="workspace.type == 'code'">
            <el-form-item label="代码库源" prop="" :required="true">
              <el-input :disabled="true" style="width: 450px;" v-model="workspace.code_url" size="small"></el-input>
            </el-form-item>
            <el-form-item label="触发分支" prop="" :required="true">
              <el-row style="margin-bottom: 5px; margin-top: 8px;">
                <el-col :span="7" style="background-color: #F5F7FA; padding-left: 10px;">
                  <div class="border-span-header">
                    <span  class="border-span-content">*</span>匹配方式
                  </div>
                </el-col>
                <el-col :span="10" style="background-color: #F5F7FA">
                  <div class="border-span-header">
                    分支
                  </div>
                </el-col>
                <!-- <el-col :span="5"><div style="width: 100px;"></div></el-col> -->
              </el-row>
              <el-row style="padding-bottom: 5px;" v-for="(d, i) in dialogData.triggers" :key="i">
                <el-col :span="7">
                  <div class="border-span-header" style="margin-right: 10px;">
                    <el-select v-model="d.operator" placeholder="匹配方式" size="small" style="width: 100%;">
                      <el-option label="精确匹配" value="equal"></el-option>
                      <el-option label="精确排除" value="exclude"></el-option>
                      <el-option label="正则匹配" value="regex"></el-option>
                    </el-select>
                  </div>
                </el-col>
                <el-col :span="10">
                  <div class="border-span-header">
                    <el-input style="border-radius: 0px;" v-model="d.branch" size="small" placeholder="匹配分支，空表示所有分支"></el-input>
                  </div>
                </el-col>
                <el-col :span="2" style="padding-left: 10px">
                  <el-button circle size="mini" style="padding: 5px;" 
                    @click="dialogData.triggers.splice(i, 1)" icon="el-icon-close"></el-button>
                </el-col>
              </el-row>
              <el-row>
                <el-col :span="17">
                <el-button style="width: 100%; border-radius: 0px; padding: 9px 15px;
                  border-color: rgb(102, 177, 255); color: rgb(102, 177, 255)" plain size="mini" 
                  @click="dialogData.triggers.push({type: 'code', branch_type: 'branch', operator: 'equal', branch: ''})" icon="el-icon-plus">添加匹配</el-button>
                </el-col>
              </el-row>
            </el-form-item>
          </el-form>

          <el-form :model="dialogData" label-position="left" label-width="105px" v-if="workspace.type == 'custom'">
            <el-form-item label="代码流水线" prop="" :required="true">
              <el-row style="margin-bottom: 5px; margin-top: 8px;">
                <el-col :span="12" style="background-color: #F5F7FA; padding-left: 10px;">
                  <div class="border-span-header">
                    <span  class="border-span-content">*</span>流水线空间
                  </div>
                </el-col>
                <el-col :span="7" style="background-color: #F5F7FA">
                  <div class="border-span-header">
                    流水线
                  </div>
                </el-col>
                <!-- <el-col :span="5"><div style="width: 100px;"></div></el-col> -->
              </el-row>
              <el-row style="padding-bottom: 5px;" v-for="(d, i) in dialogData.triggers" :key="i">
                <el-col :span="12">
                  <div class="border-span-header" style="margin-right: 10px;">
                    <el-select v-model="d.workspace" placeholder="流水线空间" size="small" style="width: 100%;">
                      <el-option v-for="w in workspaces" :key="w.id" :label="w.name" :value="w.id"></el-option>
                    </el-select>
                  </div>
                </el-col>
                <el-col :span="7">
                  <div class="border-span-header">
                    <el-select v-model="d.pipeline" placeholder="代码流水线" size="small" style="width: 100%;">
                      <el-option v-for="p in workspacesDict[d.workspace] ? workspacesDict[d.workspace].pipelines : []" :key="p.id" :label="p.name" :value="p.id"></el-option>
                    </el-select>
                  </div>
                </el-col>
                <el-col :span="2" style="padding-left: 10px">
                  <el-button circle size="mini" style="padding: 5px;" 
                    @click="dialogData.triggers.splice(i, 1)" icon="el-icon-close"></el-button>
                </el-col>
              </el-row>
              <el-row>
                <el-col :span="19">
                <el-button style="width: 100%; border-radius: 0px; padding: 9px 15px;
                  border-color: rgb(102, 177, 255); color: rgb(102, 177, 255)" plain size="mini" 
                  @click="dialogData.triggers.push({type: 'pipeline'})" icon="el-icon-plus">添加流水线源</el-button>
                </el-col>
              </el-row>
            </el-form-item>
          </el-form>
        </template>
      </div>
      <div slot="footer" class="dialogFooter">
        <el-button @click="dialogVisible = false" style="margin-right: 20px;" >取 消</el-button>
        <template v-if="dialogType == 'edit_job' || dialogType == 'edit_stage'">
        <el-button type="danger" @click="dialogDelete" style="margin-right: 20px;" >删 除</el-button>
        </template>
        <el-button type="primary" @click="dialogSave">确 定</el-button>
      </div>
    </el-dialog>

  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { PipelineStage, CodeToImage, ExecuteShell, AppDeploy, Release, DeployK8s } from '@/views/pipeline/plugin'
import { getPipeline, updatePipeline, createPipeline } from '@/api/pipeline/pipeline'
import { listWorkspaces } from '@/api/pipeline/workspace'
import { getWorkspace } from '@/api/pipeline/workspace'
import { Message } from 'element-ui'

export default {
  name: 'PipelineEdit',
  components: {
    Clusterbar,
    PipelineStage,
    CodeToImage,
    ExecuteShell,
    AppDeploy,
    Release,
    DeployK8s
  },
  data() {
    let jobPlugins = [{
      key: 'execute_shell',
      name: '执行shell脚本',
      component: 'ExecuteShell'
    }, {
      key: 'deploy_k8s',
      name: 'Kubernetes资源部署',
      component: 'DeployK8s'
    }, {
      key: 'upgrade_app',
      name: '应用部署',
      component: 'AppDeploy'
    }]
    return {
      titleName: ["流水线"],
      users: [],
      cellStyle: {border: 0, padding: '1px 0', 'line-height': '35px'},
      maxHeight: window.innerHeight - 145,
      loading: true,
      pipeline: {},
      workspace: {},
      editPipeline: {
        workspace_id: parseInt(this.$route.params.workspaceId),
        id: 0,
        name: "",
        triggers: [],
        stages: []
      },
      operatorMap: {
        "equal": "==",
        "exclude": "<>",
        "regex": "~="
      },
      dialogVisible: false,
      dialogOriginData: {},
      dialogData: {},
      dialogType: "",
      jobPlugins: jobPlugins,
      dialogTitleMap: {
        'edit_stage': '编辑阶段',
        'edit_job': '编辑任务',
        'add_stage': '添加阶段',
        'add_job': '添加任务',
        "source": '流水线源'
      },
      workspaces: [],
      workspacesDict: null,
    }
  },
  created() {
    this.fetchWorkspace()
    if(this.pipelineId) this.fetchPipeline();
    else {
      this.titleName = ["流水线", "创建"]
    }
  },
  mounted() {
    const that = this
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - 145
        console.log(heightStyle)  
        that.maxHeight = heightStyle
      })()
    }
  },
  computed: {
    workspaceId() {
      return this.$route.params.workspaceId
    },
    pipelineId() {
      return this.$route.params.pipelineId
    },
    jobPluginMap() {
      let m = {}
      for (let p of this.jobPlugins) {
        m[p.key] = p
      }
      return m
    },
    dialogTitle() {
      if(this.dialogType == 'edit_stage') return '编辑'
    }
  },
  methods: {
    fetchWorkspace() {
      this.loading = true
      getWorkspace(this.workspaceId).then((response) => {
        this.workspace = response.data || {};
        if(this.workspace.type == 'custom') {
          this.fetchWorkspaces()
        } else{
          this.jobPlugins.push({
            key: 'build_code_to_image',
            name: '构建代码镜像',
            component: 'CodeToImage'
          })
          this.jobPlugins.push({
            key: 'release',
            name: '发布',
            component: 'Release'
          })
          if(!this.pipelineId) {
            this.editPipeline.triggers = [{"type": "code", "branch_type": "branch", "operator": "equal", "branch": ""}]
          }
        }
        
        this.loading = false
      }).catch(() => {
        this.loading = false
      })
    },
    fetchPipeline() {
      this.loading = true
      getPipeline(this.pipelineId).then((response) => {
        this.pipeline = response.data || {};
        if (this.pipeline){
          this.titleName = ["流水线", this.pipeline.pipeline.name]
          this.editPipeline = {
            id: this.pipeline.pipeline.id,
            workspace_id: this.pipeline.workspace.id,
            name: this.pipeline.pipeline.name,
            triggers: this.pipeline.pipeline.triggers,
            stages: [],
          }
          for(let stage of this.pipeline.stages) {
            this.editPipeline.stages.push({
              id: stage.id,
              name: stage.name,
              trigger_mode: stage.trigger_mode,
              jobs: stage.jobs,
              custom_params: stage.custom_params,
            })
          }
        }
        this.loading = false
      }).catch(() => {
        this.loading = false
      })
    },
    savePipeline() {
      // console.log(this.pipeline)
      // console.log(this.editPipeline)
      if(!this.editPipeline.name) {
        Message.error("请输入流水线名称")
        return
      }
      this.loading = true
      if(this.pipelineId) {
        updatePipeline(this.editPipeline).then((response) => {
          Message.success("编辑流水线成功")
          this.$router.push({name: 'pipeline', params: {'workspaceId': this.workspaceId}})
          // this.loading = false
        }).catch(() => {
          this.loading = false
        })
      } else {
        createPipeline(this.editPipeline).then((response) => {
          Message.success("创建流水线成功")
          this.$router.push({name: 'pipeline', params: {'workspaceId': this.workspaceId}})
          // this.loading = false
        }).catch(() => {
          this.loading = false
        })
      }
    },
    cancelEdit() {
      this.$router.push({name: 'pipeline', params: {'workspaceId': this.workspaceId}})
    },
    dialogSave() {
      var custom_params = {}
      if(this.dialogType == 'edit_stage' || this.dialogType == 'add_stage') {
        if(this.dialogData.custom_params) {
          for(let p of this.dialogData.custom_params) {
            if(!p.param) {
              Message.error("阶段参数值不能为空")
              return
            }
            custom_params[p.param] = p.value || ''
          }
        }
      }
      if (this.dialogType == 'edit_stage') {
        this.dialogOriginData.stage.name = this.dialogData.name
        this.dialogOriginData.stage.trigger_mode = this.dialogData.trigger_mode
        this.$set(this.dialogOriginData.stage, 'custom_params', custom_params)
      } else if(this.dialogType == 'edit_job') {
        let idx = this.dialogOriginData.idx
        this.dialogOriginData.stage.jobs[idx] = this.dialogData
      } else if(this.dialogType == 'add_stage') {
        let newStage = {
          name: this.dialogData.name || '未命名',
          trigger_mode: this.dialogData.trigger_mode,
          custom_params: custom_params,
          jobs: [{
            name: "未命名",
            plugin_key: "",
            params: {},
          }]
        }
        this.editPipeline.stages.splice(this.dialogOriginData, 0, newStage)
      } else if(this.dialogType == 'add_job') {
        this.dialogOriginData.jobs.push(this.dialogData)
      } else if(this.dialogType == 'source') {
        this.editPipeline.triggers = this.dialogData.triggers
      }
      this.dialogVisible = false
    },
    dialogDelete() {
      if(this.dialogType == 'edit_stage') {
        this.editPipeline.stages.splice(this.dialogOriginData.idx, 1)
      } else if (this.dialogType == 'edit_job') {
        if(this.dialogOriginData.stage.jobs.length == 1) {
          let newJob = {
            name: "未命名",
            plugin_key: "",
            params: {},
          }
          this.dialogOriginData.stage.jobs[0] = newJob
        } else {
          this.dialogOriginData.stage.jobs.splice(this.dialogOriginData.idx, 1)
        }
      }
      this.dialogVisible = false
    },
    openEditStageDialog(stage, idx) {
      this.dialogType = 'edit_stage'
      this.dialogOriginData = {
        stage,
        idx
      }
      console.log(stage)
      var custom_params = []
      if(stage.custom_params) {
        for(let k in stage.custom_params) {
          custom_params.push({param: k, value: stage.custom_params[k]})
        }
      }
      this.dialogData = {
        name: stage.name,
        trigger_mode: stage.trigger_mode,
        custom_params: custom_params
      }
    },
    openEditJobDialog(stage, idx) {
      this.dialogType = 'edit_job'
      this.dialogOriginData = {
        stage: stage,
        idx: idx,
      }
      this.dialogData = JSON.parse(JSON.stringify(stage.jobs[idx]))
    },
    openAddStageDialog(idx) {
      this.dialogType = 'add_stage'
      this.dialogOriginData = idx
      this.dialogData = {
        name: "",
        trigger_mode: "auto",
        custom_params: []
      }
    },
    openAddJobDialog(stage) {
      this.dialogType = 'add_job'
      this.dialogOriginData = stage
      this.dialogData = {
        name: "未命名",
        plugin_key: "",
        params: {},
      }
    },
    checkJobError(job) {
      if(job.name == '') return false
      if(job.plugin_key == '') return false
      return true
    },
    openEditSource() {
      this.dialogType = 'source'
      this.dialogData = {
        triggers: JSON.parse(JSON.stringify(this.editPipeline.triggers))
      }
      this.dialogVisible = true
    },
    fetchWorkspaces() {
      listWorkspaces({"with_pipeline": true, "type": "code"})
        .then((response) => {
          this.workspacesDict = {}
          this.workspaces = response.data || [];
          for(let w of this.workspaces) {
            this.workspacesDict[w.id] = w
          }
        }).catch(() => {})
    },
    getTriggerWorkspaceName(workspaceId) {
      if(this.workspacesDict) {
        if(this.workspacesDict[workspaceId]) {
          return this.workspacesDict[workspaceId].name
        } 
        return '未知'
      }else {
        return ''
      }
    },
    getTriggerPipelineName(workspaceId, pipelineId) {
      if(this.workspacesDict) {
        if(this.workspacesDict[workspaceId]) {
          for(let p of this.workspacesDict[workspaceId].pipelines) {
            if(p.id == pipelineId) {
              return p.name
            }
          }
        } 
        return '未知'
      }else {
        return ''
      }
    }
  }
}
</script>

<style lang="scss" scoped>
.bar-btn {
  padding: 9px 25px
}

.stage-job-outer {
  padding: 10px 20px;
  display: flex; 
  height: 500px; 
  max-height: 500px; 
  overflow-x: scroll;

  .stage-job-line {
    display:inline-block;
  }

  .stage-job-block {
    font-size: 14px;
    width: 269px;

    .stage-job-block__stage {
      text-align: center; 
      display: inline-block;
      border-radius: 15px; 
      padding: 8px 15px; 
      margin: 0px 0px 10px 30px; 
      background-color: #EBEEF5;
      height: 32px;
    }

    .stage-job-block__stage:hover{
      cursor: pointer;
      color: #409EFF;
    }

    .stage-job-block__job-name {
      width: 250px; 
      text-align: center; 
      color: #606266; 
      margin-top: 5px;
    }

    .stage-job-block__job {
      width: 210px; 
      height: 80px; 
      text-align: center; 
      border-left: 1px solid #c0c4cc; 
      border-right: 1px solid #c0c4cc; 
      border-bottom: 1px solid #c0c4cc;  
      margin: -35px 20px 0px;

      .stage-job-block__job-circle {
        font-size: 14px;
        line-height: 25px; 
        width: 210px;
        text-align: center;
        padding-top: 66px;
      }
    }

    .stage-job-block__job-add {
      width: 210px; 
      height: 80px; 
      text-align: center; 
      border-left: 1px dashed #c0c4cc; 
      border-right: 1px dashed #c0c4cc; 
      border-bottom: 1px dashed #c0c4cc;  
      margin: -35px 20px;

      .stage-job-block__job-add__inner {
        font-size: 14px;
        line-height: 25px; 
        width: 210px;
        text-align: center;
        padding-top: 66px;
      }

      .stage-job-block__job-add__inner-name {
        background-color: white;
        margin-top: 66px; 
        border-radius: 15px; 
        border:1px dashed #C0C4CC; 
        padding: 4px 8px; 
        color: #909399
      }

      .stage-job-block__job-add__inner-name:hover{
        cursor: pointer;
      }
    }
  }

  .stage-job-line__inner {
    height: 3px; 
    width: 113px; 
    background-color: #c0c4cc; 
    display: inline-block; 
    vertical-align: middle
  }

  .stage-job-line__circle{
    background-color: white;
    height: 25px; 
    width: 25px; 
    border: 3px solid #c0c4cc; 
    display: inline-block; 
    vertical-align: middle; 
    border-radius: 50%;
  }

  .stage-job-line__circle-error {
    background-color: white;
    height: 25px; 
    width: 25px; 
    border: 3px solid #ec7676; 
    display: inline-block; 
    vertical-align: middle; 
    border-radius: 50%;
    text-align: center;
    color: #ec7676;
    font-size: 16px;
    font-weight: 500;
  }

  .stage-job-line__circle-error:hover {
    cursor: pointer;
    border-color: #ec7676;
    background-color: #ec7676;
  }

  .stage-job-line__circle:hover {
    cursor: pointer;
    border-color: #409EFF;
    background-color: #409EFF;
  }

  .stage-job-line__add {
    font-size: 20px;
    color: #606266;
    display: inline-block;
    vertical-align: middle; 
    margin-left: -1px; 
    margin-right: -1px;
  }

  .stage-job-line__add:hover{
    cursor: pointer;
    color: #409EFF;
  }

  .pipeline-source-outer {
    border:1px solid #c0c4cc; 
    border-radius: 12px; 
    padding: 12px 8px; 
    margin-top: 20px;

    .pipeline-source-outer__span{
      border: 1px solid rgb(140, 197, 255); 
      background-color: rgb(236, 245, 255); 
      border-radius: 25px; 
      padding: 4px 8px; 
      font-size: 12px; 
      color: #409EFF;
    }
  }
  .pipeline-source-outer:hover{
    border-color: #409EFF;
    cursor: pointer;
  }

}

</style>

<style lang="scss">
.pipeline-form-item {
  padding: 10px 20px;
  font-size: 0;
  
  label {
    width: 90px;
    color: #99a9bf;
    font-weight: 400;
  }
 .el-form-item {
    margin-right: 0;
    margin-bottom: 0;
    width: 30%;
  }
  input {
    border-radius: 0px;
  } 
}

</style>