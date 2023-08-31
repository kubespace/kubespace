<template>
  <div>
    <clusterbar :titleName="titleName" :titleLink="['pipeline']">
      <div slot="right-btn">
        <el-button size="small" class="bar-btn" type="" @click="cancelEdit">取消</el-button>
        <el-button size="small" class="bar-btn" type="primary" @click="savePipeline">保存</el-button>
      </div>
    </clusterbar>
    <div v-loading="loading" class="dashboard-container">
      <div style="padding: 10px 0px 0px;">
        <div>
          基本信息
        </div>
        <el-form label-position="left" class="pipeline-form-item" label-width="80px">
          <el-form-item label="名称">
            <el-input v-model="editPipeline.name" autocomplete="off" placeholder="请输入流水线名称" size="small"></el-input>
          </el-form-item>
        </el-form>
      </div>
      <div style="padding: 10px 0px 0px;" >
        <div>阶段任务</div>
        <div class="stage-job-outer" :style="{height: maxHeight + 'px'}">
          <div class="stage-job-line">
            <div class="pipeline-source-outer" style="display:inline-block" @click="openEditSource()">
              <div v-if="workspace.type == 'code'">
                <span class="pipeline-source-outer__span">
                  代码库源
                </span>
                <div style="font-size: 12px; padding: 10px 0px 0px; font-weight: 450">
                  {{ workspace ? workspace.code ? workspace.code.clone_url : '' : '' }}
                </div>
                <div v-for="(t, i) in editPipeline.sources" :key="i"
                  style="font-size: 12px; padding: 5px 10px 0px; font-weight: 400" >
                  <svg-icon icon-class="branch" /> {{ operatorMap[t.operator] }} {{ t.branch }}
                </div>
              </div>
              <div v-else>
                <span class="pipeline-source-outer__span">
                  流水线源
                </span>
                <template v-if="editPipeline.sources && editPipeline.sources.length > 0">
                  <div v-for="(t, i) in editPipeline.sources" :key="i">
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
          <div >
            <div style="margin-top: 43px; width: 38px; display: inline-block;">
              <div class="stage-job-line__inner" style="width: 20px;"></div>
              <div class="stage-job-line__add" @click="openAddStageDialog(0); dialogVisible=true;">
                <el-tooltip class="item" effect="light" content="添加阶段" placement="top" :hide-after="2000">
                  <i class="el-icon-circle-plus"></i>
                </el-tooltip>
              </div>
            </div>
          </div>
          <div v-for="(stage, i) in editPipeline.stages" :key="i" style="">
            <div class="stage-job-line" style="margin-top: 38px;">
              <div style="display: inline-flex;">
                <div>
                  <div class="stage-job-line__inner" style="width: 30px;" ></div>
                  <div class="stage-job-block" style="display: inline-block;">
                    <div class="stage-job-block__stage" @click="openEditStageDialog(stage, i); dialogVisible=true;">
                      <span>#{{ i + 1 }} {{ stage.name }}</span>
                    </div>
                  </div>
                  <div class="stage-job-line__inner" style="width: 30px;"></div>
                  <div class="stage-job-line__add" @click="openAddStageDialog(i+1); dialogVisible=true;">
                    <el-tooltip class="item" effect="light" content="添加阶段" placement="top" :hide-after="2000">
                      <i class="el-icon-circle-plus"></i>
                    </el-tooltip>
                  </div>
                </div>
              </div>
              <div class="stage-job-block" style="margin-left: 25px; ">
                <template  v-if="stage.jobs" >
                  <div v-for="(job, ji) in stage.jobs" :key="ji">
                      <div class="stage-job-block__job">
                        <div class="stage-job-block__job-circle">
                            <span class="stage-job-block__job-add__inner-name stage-job-block__job-add__inner-name-hover" 
                              @click="openEditJobDialog(stage, ji); dialogVisible=true;">{{ job.name }}</span>
                        </div>
                      </div>
                  </div>
                </template>

                <div class="stage-job-block__job-add">
                  <div class="stage-job-block__job-add__inner">
                    <span class="stage-job-block__job-add__inner-name stage-job-block__job-add__inner-ex" 
                    @click="openAddJobDialog(stage); dialogVisible=true;">+ 新建并行任务</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

        </div>
      </div>
    </div>

    <el-drawer :title="dialogTitleMap[dialogType]" :visible.sync="dialogVisible" :destroy-on-close="true" :wrapperClosable="false"
      @close="closeDialog()" top="3vh" size="60%" :close-on-click-modal="false" style="scroll-behavior: auto;">
      <div slot="title">
        <span>{{ dialogTitleMap[dialogType] }}</span>
        <div style="display: inline; ">
          <template v-if="dialogType == 'edit_job' || dialogType == 'edit_stage'">
            <el-link type="danger" icon="el-icon-delete" @click="dialogDelete" style="margin-left: 5px; font-size: 18px" ></el-link>
          </template>
        </div>
      </div>
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
              <div class="job-advanced" v-if="hasAdvancedPlugins(dialogData.plugin_key)">
                <el-divider>
                  <div class="job-advanced-text" @click="turnJobAdvanced">
                    <i :class="jobAdvancedVisible ? 'el-icon-caret-bottom' : 'el-icon-caret-right'" style="margin-right: 5px;"></i>高级设置
                  </div>
                </el-divider>
                <div v-if="jobAdvancedVisible">
                  <el-form-item label="指定Spacelet节点" prop="" label-width="145px">
                    <div slot="label">
                      指定Spacelet节点
                      <el-popover placement="top-start" title="" width="500" trigger="hover">
                        <div style="line-height: 24px;">
                            任务执行时调度到该Spacelet节点
                        </div>
                        <i slot="reference" class="el-icon-question"></i>
                      </el-popover>
                    </div>
                    <el-select v-model="dialogData.schedule_policy.hostname" placeholder="请选择要指定的Spacelet节点" size="small" style="width: 330px">
                      <el-option :key="-1" label="不选择" value=""></el-option>
                      <el-option :disabled="res.status!='online'"
                        v-for="res in spacelets"
                        :key="res.hostname"
                        :label="res.hostname"
                        :value="res.hostname">
                      </el-option>
                    </el-select>
                  </el-form-item>
                  <el-form-item label="Spacelet标签选择" prop="" label-width="145px">
                    <div slot="label">
                      Spacelet标签选择
                      <el-popover placement="top-start" title="" width="500" trigger="hover">
                        <div style="line-height: 24px;">
                            任务执行时调度到匹配标签的Spacelet节点
                        </div>
                        <i slot="reference" class="el-icon-question"></i>
                      </el-popover>
                    </div>
                    <div style="margin-bottom: 5px;" v-for="(l, i) in dialogData.schedule_policy.spacelet_selector" :key="i">
                      <el-input size="small" v-model="l.key" style="width: 25%;" placeholder="Key"></el-input> = 
                      <el-input size="small" v-model="l.value" style="width: 25%;" placeholder="Value"></el-input>
                      <el-button size="mini" circle style="padding: 5px; margin-left: 10px;" @click="dialogData.schedule_policy.spacelet_selector.splice(i, 1)" 
                        icon="el-icon-close"></el-button>
                    </div>
                    <el-button plain size="small" @click="dialogData.schedule_policy.spacelet_selector.push({key: '', value: ''})" icon="el-icon-plus"
                      style="border-radius: 0px;">添加</el-button>
                  </el-form-item>
                </div>
              </div>
            </div>
          </el-form>
        </template>
        <template v-if="dialogType == 'source'">
          <el-form :model="dialogData" label-position="left" label-width="105px">
            <template v-if="workspace.type == 'code'">
              <el-form-item label="代码库源" prop="" :required="true">
                <el-input :disabled="true" style="width: 450px;" v-model="workspace.code.clone_url" size="small"></el-input>
              </el-form-item>
              <el-form-item label="构建分支" prop="" :required="true">
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
                </el-row>
                <el-row style="padding-bottom: 5px;" v-for="(d, i) in dialogData.sources" :key="i">
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
                      @click="dialogData.sources.splice(i, 1)" icon="el-icon-close"></el-button>
                  </el-col>
                </el-row>
                <el-row>
                  <el-col :span="17">
                  <el-button style="width: 100%; border-radius: 0px; padding: 9px 15px; border-color: rgb(102, 177, 255); color: rgb(102, 177, 255)" plain size="mini" 
                    @click="dialogData.sources.push({type: 'code', branch_type: 'branch', operator: 'equal', branch: ''})" icon="el-icon-plus">添加匹配</el-button>
                  </el-col>
                </el-row>
              </el-form-item>
              <el-form-item label="自动触发" prop="">
                <div slot="label">
                  自动触发
                  <span>
                    <el-popover placement="top-start" title="" width="500" trigger="hover">
                      <div style="line-height: 20px;">
                          代码提交到当前流水线匹配的分支后，是否自动触发该流水线
                      </div>
                      <i slot="reference" class="el-icon-question"></i>
                    </el-popover>
                  </span>
                </div>
                <el-switch v-model="dialogData.code_trigger" ></el-switch>
              </el-form-item>
            </template>

            <template v-if="workspace.type == 'custom'">
              <el-form-item label="代码流水线" prop="" :required="true">
                <el-row style="margin-bottom: 5px; margin-top: 8px;">
                  <el-col :span="12" style="background-color: #F5F7FA; padding-left: 10px;">
                    <div class="border-span-header">
                      <span  class="border-span-content">*</span>流水线空间
                    </div>
                  </el-col>
                  <el-col :span="8" style="background-color: #F5F7FA">
                    <div class="border-span-header">
                      流水线
                    </div>
                  </el-col>
                  <!-- <el-col :span="5"><div style="width: 100px;"></div></el-col> -->
                </el-row>
                <el-row style="padding-bottom: 5px;" v-for="(d, i) in dialogData.sources" :key="i">
                  <el-col :span="12">
                    <div class="border-span-header" style="margin-right: 10px;">
                      <el-select v-model="d.workspace" placeholder="流水线空间" size="small" style="width: 100%;" @change="changePipeline">
                        <el-option v-for="w in workspaces" :key="w.id" :label="w.name" :value="w.id"></el-option>
                      </el-select>
                    </div>
                  </el-col>
                  <el-col :span="8">
                    <div class="border-span-header">
                      <el-select v-model="d.pipeline" placeholder="代码流水线" size="small" style="width: 100%;">
                        <el-option v-for="p in workspacesDict[d.workspace] ? workspacesDict[d.workspace].pipelines : []" :key="p.id" :label="p.name" :value="p.id"></el-option>
                      </el-select>
                    </div>
                  </el-col>
                  <el-col :span="2" style="padding-left: 10px">
                    <el-button circle size="mini" style="padding: 5px;" 
                      @click="dialogData.sources.splice(i, 1)" icon="el-icon-close"></el-button>
                  </el-col>
                </el-row>
                <el-row>
                  <el-col :span="20">
                  <el-button style="width: 100%; border-radius: 0px; padding: 9px 15px; border-color: rgb(102, 177, 255); color: rgb(102, 177, 255)" plain size="mini" 
                    @click="dialogData.sources.push({type: 'pipeline'})" icon="el-icon-plus">添加流水线源</el-button>
                  </el-col>
                </el-row>
              </el-form-item>
            </template>
              <el-form-item label="定时触发" prop="">
                <el-switch v-model="dialogData.cron_trigger" ></el-switch>
                <div v-if="dialogData.cron_trigger">
                  <!-- <el-input v-model="dialogData.cron" size="small" placeholder="定时策略"></el-input> -->
                  <el-row style="margin-bottom: 5px; margin-top: 8px;">
                    <el-col :span="3" style="background-color: #F5F7FA; padding-left: 10px;">
                      <div class="border-span-header">
                        <span  class="border-span-content">*</span>分
                      </div>
                    </el-col>
                    <el-col :span="3" style="background-color: #F5F7FA">
                      <div class="border-span-header">
                        <span  class="border-span-content">*</span>时
                      </div>
                    </el-col>
                    <el-col :span="3" style="background-color: #F5F7FA">
                      <div class="border-span-header">
                        <span  class="border-span-content">*</span>日
                      </div>
                    </el-col>
                    <el-col :span="3" style="background-color: #F5F7FA">
                      <div class="border-span-header">
                        <span  class="border-span-content">*</span>月
                      </div>
                    </el-col>
                    <el-col :span="3" style="background-color: #F5F7FA">
                      <div class="border-span-header">
                        <span  class="border-span-content">*</span>周
                      </div>
                    </el-col>
                    <!-- <el-col :span="5"><div style="width: 100px;"></div></el-col> -->
                  </el-row>
                  <el-row style="padding-bottom: 5px;">
                    <el-col :span="3" style="padding-right: 10px;">
                      <div class="border-span-header" >
                        <el-input v-model="dialogData.cron_min" size="small"></el-input>
                      </div>
                    </el-col>
                    <el-col :span="3" style="padding-right: 10px;">
                      <div class="border-span-header">
                        <el-input v-model="dialogData.cron_hour" size="small"></el-input>
                      </div>
                    </el-col>
                    <el-col :span="3" style="padding-right: 10px;">
                      <div class="border-span-header">
                        <el-input v-model="dialogData.cron_day" size="small"></el-input>
                      </div>
                    </el-col>
                    <el-col :span="3" style="padding-right: 10px;">
                      <div class="border-span-header">
                        <el-input v-model="dialogData.cron_mon" size="small"></el-input>
                      </div>
                    </el-col>
                    <el-col :span="3" style="padding-right: 10px;">
                      <div class="border-span-header">
                        <el-input v-model="dialogData.cron_week" size="small"></el-input>
                      </div>
                    </el-col>
                  </el-row>
                </div>
              </el-form-item>
          </el-form>
        </template>
        <div style="display: block; padding: 25px 0px; text-align: center;">
          <el-button type="primary" @click="dialogSave" style="margin-right: 25px;">保 存</el-button>
          <el-button @click="dialogVisible = false">取 消</el-button>
        </div>
      </div>
    </el-drawer>

  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { PipelineStage, CodeToImage, ExecuteShell, AppDeploy, Release, DeployK8s, checkPluginJob } from '@/views/pipeline/plugin'
import { getPipeline, updatePipeline, createPipeline } from '@/api/pipeline/pipeline'
import { listWorkspaces } from '@/api/pipeline/workspace'
import { getWorkspace } from '@/api/pipeline/workspace'
import { listSpacelet } from "@/api/spacelet/spacelet";
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
      maxHeight: window.innerHeight - 295,
      loading: true,
      pipeline: {},
      workspace: {},
      editPipeline: {
        workspace_id: parseInt(this.$route.params.workspaceId),
        id: 0,
        name: "",
        sources: [],
        stages: [],
        triggers: [],
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

      jobAdvancedVisible: false,
      spacelets: [],
      spaceletsLoaded: false,
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
        let heightStyle = window.innerHeight - 295
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
    },
    originCodeTrigger() {
      if(!this.pipeline) {
        return null
      }
      if(!this.pipeline.triggers) {
        return null
      }
      for(let t of this.pipeline.triggers) {
        if(t.type == "code") {
          return t
        }
      }
      return null
    },
  },
  methods: {
    hasAdvancedPlugins(pluginKey) {
      return ['build_code_to_image', 'release', 'execute_shell'].indexOf(pluginKey) >= 0
    },
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
            this.editPipeline.sources = [{"type": "code", "branch_type": "branch", "operator": "equal", "branch": ""}]
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
            sources: this.pipeline.pipeline.sources,
            stages: [],
          }
          if(this.pipeline.triggers) {
            this.editPipeline.triggers = JSON.parse(JSON.stringify(this.pipeline.triggers))
          } else {
            this.editPipeline.triggers = []
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
      if(!this.editPipeline.stages || this.editPipeline.stages.length == 0) {
        Message.error("流水线阶段不能为空")
        return
      }
      for(let s of this.editPipeline.stages) {
        if(!s.jobs || s.jobs.length == 0 ) {
          Message.error(`流水线阶段"${s.name}"任务不能为空`)
          return
        }
      }
      this.loading = true
      if(this.pipelineId) {
        updatePipeline(this.pipelineId, this.editPipeline).then((response) => {
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
      switch(this.dialogType) {
        case 'edit_stage':
        case 'add_stage':
          var custom_params = {}
          if(this.dialogData.custom_params) {
            for(let p of this.dialogData.custom_params) {
              if(!p.param) {
                Message.error("阶段参数不能为空")
                return
              }
              custom_params[p.param] = p.value || ''
            }
          }
          if (this.dialogType == 'edit_stage') {
            this.dialogOriginData.stage.name = this.dialogData.name
            this.dialogOriginData.stage.trigger_mode = this.dialogData.trigger_mode
            this.$set(this.dialogOriginData.stage, 'custom_params', custom_params)
          } else {
            let newStage = {
              name: this.dialogData.name || '未命名',
              trigger_mode: this.dialogData.trigger_mode,
              custom_params: custom_params,
              jobs: [],
              // jobs: [{
              //   name: "未命名",
              //   plugin_key: "",
              //   params: {},
              // }]
            }
            this.editPipeline.stages.splice(this.dialogOriginData, 0, newStage)
          }
          break
        
        case "edit_job":
        case "add_job":
          let checkRes = checkPluginJob(this.dialogData)
          if(!checkRes.checked) {
            Message.error(checkRes.errorMsg)
            return
          }
          if(this.dialogData.schedule_policy) {
            let schedulePolicy = this.dialogData.schedule_policy
            if(schedulePolicy.spacelet_selector) {
              let selector = {}
              for(let s of schedulePolicy.spacelet_selector) {
                if(!s.key) {
                  Message.error("Spacelet标签选择键不能为空")
                  return
                }
                if(s.key in selector) {
                  Message.error(`Spacelet标签选择键「${s.key}」重复`)
                  return
                }
                selector[s.key] = s.value
              }
              schedulePolicy.spacelet_selector = selector
            }
          }
          if(this.dialogType == 'edit_job') {
            let idx = this.dialogOriginData.idx
            this.dialogOriginData.stage.jobs[idx] = this.dialogData
          } else {
            this.dialogOriginData.jobs.push(this.dialogData)
          }
          break

        case "source":
          this.$set(this.editPipeline, 'sources', this.dialogData.sources)
          let hasCodeTrigger = this.hasCodeTrigger()
          if(this.dialogData.code_trigger && !hasCodeTrigger) {
            if(this.originCodeTrigger) {
              this.editPipeline.triggers.push(this.originCodeTrigger)
            } else {
              this.editPipeline.triggers.push({type: "code"})
            }
          } else if(!this.dialogData.code_trigger && hasCodeTrigger) {
            let triggers = []
            for(let t of this.editPipeline.triggers) {
              if (t.type != "code") {
                triggers.push(t)
              }
            }
            this.editPipeline.triggers = triggers
          }
          let originCron = this.originCronTrigger()
          let editCron = this.editCronTrigger()
          if (this.dialogData.cron_trigger) {
            let cronStr = this.getCron()
            if(editCron) {
              editCron['cron'] = cronStr
            } else {
              let cron = {type: "cron", cron: cronStr}
              if(originCron) {
                cron["id"] = originCron.id
              }
              this.editPipeline.triggers.push(cron)
            }
          } else if(!this.dialogData.cron_trigger && editCron) {
            let triggers = []
            for(let t of this.editPipeline.triggers) {
              if (t.type != "cron") {
                triggers.push(t)
              }
            }
            this.editPipeline.triggers = triggers
          }
          break
      }
      
      this.dialogVisible = false
    },
    getCron() {
      let crons = []
      for(let i of ["cron_min", "cron_hour", "cron_day", "cron_mon", "cron_week"]) {
        if(this.dialogData[i]) {
          crons.push(this.dialogData[i])
        } else {
          crons.push('*')
        }
      }
      return crons.join(' ')
    },
    dialogDelete() {
      if(this.dialogType == 'edit_stage') {
        this.editPipeline.stages.splice(this.dialogOriginData.idx, 1)
      } else if (this.dialogType == 'edit_job') {
        this.dialogOriginData.stage.jobs.splice(this.dialogOriginData.idx, 1)
      }
      this.dialogVisible = false
    },
    openEditStageDialog(stage, idx) {
      this.dialogType = 'edit_stage'
      this.dialogOriginData = {
        stage,
        idx
      }
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
      // console.log(this.dialogData)
      let schedulePolicy = this.dialogData.schedule_policy
      if (!schedulePolicy) {
        this.$set(this.dialogData, 'schedule_policy', {hostname: '', spacelet_selector: []})
      } else if(!schedulePolicy.spacelet_selector){
        this.$set(schedulePolicy, 'spacelet_selector', [])
      } else if(schedulePolicy.spacelet_selector) {
        let selectors = []
        for(let sk in schedulePolicy.spacelet_selector) {
          selectors.push({key: sk, value: schedulePolicy.spacelet_selector[sk]})
        }
        this.$set(schedulePolicy, 'spacelet_selector', selectors)
      }
      // console.log(this.dialogData)
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
        schedule_policy: {hostname: '', spacelet_selector: []},
      }
    },
    checkJobError(job) {
      if(job.name == '') return false
      if(job.plugin_key == '') return false
      return true
    },
    openEditSource() {
      this.dialogType = 'source'
      // this.dialogData = {
      //   sources: JSON.parse(JSON.stringify(this.editPipeline.sources))
      // }
      
      this.$set(this.dialogData, 'sources', JSON.parse(JSON.stringify(this.editPipeline.sources)))
      this.$set(this.dialogData, "code_trigger", this.hasCodeTrigger())
      let editCron = this.editCronTrigger()
      this.$set(this.dialogData, "cron_trigger", !!editCron)
      let crons = ['*', '*', '*', '*', '*']
      if(editCron) {
        crons = this.splitCron(editCron.cron)
      }
      this.$set(this.dialogData, "cron_min", crons[0])
      this.$set(this.dialogData, "cron_hour", crons[1])
      this.$set(this.dialogData, "cron_day", crons[2])
      this.$set(this.dialogData, "cron_mon", crons[3])
      this.$set(this.dialogData, "cron_week", crons[4])
      
      this.dialogVisible = true
    },
    splitCron(cronStr) {
      let crons = cronStr.split(" ")
      for(let i in crons) {
        if(!crons[i]) {
          crons[i] = '*'
        }
      }
      return crons
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
    },
    changePipeline(val) {
      // this.$set(p, 'pipeline', '')
      for(let d of this.dialogData.sources) {
        if(d.workspace == val) {
          if(d.pipeline) delete d.pipeline
          if(d.workspace_name) d.workspace_name = ''
          if(d.pipeline_name) d.pipeline_name = ''
        }
      }
    },
    originCronTrigger() {
      if(!this.pipeline) {
        return null
      }
      if(!this.pipeline.triggers) {
        return null
      }
      for(let t of this.pipeline.triggers) {
        if(t.type == "cron") {
          return t
        }
      }
      return null
    },
    editCronTrigger() {
      if(!this.editPipeline) return null
      if(!this.editPipeline.triggers) {
        return null
      }
      for(let t of this.editPipeline.triggers) {
        if (t.type == "cron") {
          return t
        }
      }
      return null
    },
    hasCodeTrigger() {
      if(!this.editPipeline) return false
      if(!this.editPipeline.triggers) {
        return false
      }
      for(let t of this.editPipeline.triggers) {
        if (t.type == "code") {
          return true
        }
      }
      return false
    },
    turnJobAdvanced() {
      this.jobAdvancedVisible = !this.jobAdvancedVisible
      if(!this.jobAdvancedVisible) return
      if(!this.spaceletsLoaded) {
        listSpacelet().then((resp) => {
          this.spacelets = resp.data ? resp.data : []
          this.spaceletsLoaded = true
        }).catch((err) => {
          console.log(err)
        })
      }
    },
    closeDialog() {
      this.dialogType=''
      this.dialogData={}
      this.jobAdvancedVisible=false
      // this.dialogVisible = false
    }
  }
}
</script>

<style lang="scss" scoped>
.bar-btn {
  padding: 9px 25px;
}

.stage-job-outer {
  padding: 10px 20px;
  display: -webkit-box;
  // height: 500px;
  // max-height: 500px;
  overflow: auto;
  background-color: #fff;
  border-radius: 10px;
  margin: 15px 0px 0px;

  .stage-job-line {
    display:inline-block;
  }

  .stage-job-block {
    font-size: 14px;
    //width: 269px;

    .stage-job-block__stage {
      text-align: center;
      display: inline-block;
      border-radius: 10px;
      padding: 8px 15px;
      background-color: #EBEEF5;
      height: 32px;
      min-width: 120px;
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
      width: 20px;
      height: 40px;
      border-left: 1px solid #c0c4cc;
      border-bottom: 1px solid #c0c4cc;
      margin: 0px 20px 0px;

      .stage-job-block__job-circle {
        font-size: 14px;
        line-height: 25px;
        width: 210px;
        padding-top: 26px;
      }
    }
    .stage-job-block__job-add__inner-name-hover:hover {
      border-color: #409EFF;
      color: #409EFF;
    }
    .stage-job-block__job-add__inner-name {
      margin-left: 20px;
      background-color: white;
      margin-top: 36px; 
      border-radius: 15px; 
      border:1px solid #C0C4CC;
      padding: 4px 8px;
    }

    .stage-job-block__job-add__inner-name:hover{
      cursor: pointer;
    }

    .stage-job-block__job-add {
      width: 20px;
      height: 40px;
      border-left: 1px dashed #c0c4cc;
      border-bottom: 1px dashed #c0c4cc;
      margin: 0px 20px;
      opacity: 0.6;

      .stage-job-block__job-add__inner {
        font-size: 14px;
        line-height: 25px;
        width: 210px;
        padding-top: 26px;

        .stage-job-block__job-add__inner-ex {
          border:1px dashed #C0C4CC;
          color: #909399;
        }
        .stage-job-block__job-add__inner-ex:hover {
          border-color: #409EFF;
          color: #409EFF;
        }
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
  background-color: #fff;
  border-radius: 10px;
  margin: 15px 0px 5px;
  
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

.job-advanced {
  .el-divider__text {
    background-color: #F2F6FC;
  }
  .job-advanced-text:hover {
    cursor: pointer;
  }
}

</style>