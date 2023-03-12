<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" createDisplay="构建" :titleLink="['pipeline']">
      <el-button  slot="right-btn" size="small" type="primary" @click="openBuildParams">
        <i class="el-icon-video-play" style=""></i> 构 建
      </el-button>
    </clusterbar>
    <div v-loading="loading" class="dashboard-container dashboard-container-build" :style="{'max-height': maxHeight + 'px'}" :max-height="maxHeight">
      <div class="build-list" v-for="build in builds || []" :key="build.pipeline_run.id">
        <div style="border-bottom: 1px solid #EBEEF5;">
          <div class="build-info">
            <div class="build-info__left" @click="clickBuildDetail(build, 'source')">
              <div class="build-info__left-number">
                <div class="build-info__left-number-inner" 
                  :style="{color: getBuildStatusColor(build.pipeline_run.status), 
                    'line-height': pipeline.workspace.type == 'code' ? '22px' : '45px'}">
                  <status-icon :status="build.pipeline_run.status"></status-icon>
                  <span class="build-info__left__number" @click="clickBuildNumber(build)"> #{{ build.pipeline_run.build_number }}</span>
                  <!-- #{{ build.pipeline_run.build_number }} -->
                </div>
                <div class="build-info__left-branch" v-if="pipeline.workspace && pipeline.workspace.type == 'code'">
                  <svg-icon icon-class="branch" /> {{ build.pipeline_run.env.PIPELINE_CODE_BRANCH }}
                </div>
              </div>
              <div class="build-info__left-op">
                <div style="height: 22px; line-height: 22px; vertical-align: middle;">
                  <i class="el-icon-user"></i> {{ build.pipeline_run.operator }}
                </div>
                <div style="font-size: 12px; line-height: 22px; height: 22px; vertical-align: middle;">
                  <i class="el-icon-date"></i> {{ $dateFormat(build.pipeline_run.create_time) }}
                </div>
              </div>
            </div>
            <el-row style="width: 100%;">
              <el-steps simple class="el-steps">
                <el-step title="" icon="none" :status="getStageStatus(stage.status)" v-for="stage in build.stages_run" :key="stage.id">
                  <div slot="title" class="el-steps-title">
                    <div><span class="el-steps-title-name" style="margin-left: 1px; overflow: hidden; white-space: nowrap;"  @click="clickBuildDetail(build, 'stage', stage)">{{ stage.name }}{{ releaseVersion(stage) }}</span> </div>
                    <div style="margin-top: 3px;">
                      <template v-if="stage.status == 'ok'">
                        <i class="el-icon-circle-check" style="font-size: 18px;"></i>
                        <div class="el-steps-stage-exectime">
                          {{ getStageExecTime(stage.exec_time, stage.update_time) }}
                        </div>
                      </template>
                      <template v-if="stage.status == 'doing'">
                        <i class="el-icon-refresh refresh-rotate" style="font-size: 18px;"></i>
                        <div class="el-steps-stage-exectime">
                          {{ getStageExecTimeStr(refreshStages[stage.id]) }}
                        </div>
                      </template>
                      <template v-if="stage.status == 'error'">
                        <i class="el-icon-circle-close" style="font-size: 18px;"></i>
                        <div class="el-steps-stage-exectime">
                          {{ getStageExecTime(stage.exec_time, stage.update_time) }}
                        </div>
                      </template>
                      <template v-if="stage.status == 'wait'">
                        <i class="el-icon-remove-outline" style="font-size: 18px;"></i>
                        <div class="el-steps-stage-exectime">
                          --
                        </div>
                      </template>
                      <template v-if="stage.status == 'cancel'">
                        <i class="el-icon-remove-outline" style="font-size: 18px;"></i>
                        <div class="el-steps-stage-exectime">
                          {{ getStageExecTime(stage.exec_time, stage.finish_time) }}
                        </div>
                      </template>
                      <template v-if="stage.status == 'canceled'">
                        <i class="el-icon-remove-outline" style="font-size: 18px;"></i>
                        <div class="el-steps-stage-exectime">
                          {{ getStageExecTime(stage.exec_time, stage.finish_time) }}
                        </div>
                      </template>
                      <template v-if="stage.status == 'pause'">
                        <div>
                          <el-button size="mini" type="primary" style="padding: 2px 5px;" round
                            @click="openManualStageDialog(stage)">
                            <i class="el-icon-video-play" /> 
                            执行
                          </el-button>
                        </div>
                      </template>
                    </div>
                  </div>
                </el-step>
              </el-steps>
            </el-row>
          </div>
          <div class="build-detail" v-if="build.clickDetail"
            style="border-top: 1px solid #EBEEF5; padding: 10px 15px 15px;">
            <template v-if="build.clickDetail && build.clickDetail.type == 'source'">
              <div style="font-size: 14px; padding: 4px 3px 8px">
                构建源
              </div>
              <el-table v-if="pipeline.workspace.type == 'code'"
                :data="build.clickDetail ? build.clickDetail.commit || [] : []"
                :cell-style="cellStyle"
                style="width: 100%">
                <el-table-column
                  prop="commitId"
                  label="CommitId"
                  width="330">
                </el-table-column>
                <el-table-column
                  prop="author"
                  label="Author"
                  width="100">
                </el-table-column>
                <el-table-column
                  prop="when"
                  label="When"
                  width="160">
                  <template slot-scope="scope">
                    {{ scope.row.when ? $dateFormat(scope.row.when) : "" }}
                  </template>
                </el-table-column>
                <el-table-column
                  prop="message"
                  label="Comment"
                  width="">
                </el-table-column>
              </el-table>
              <el-table v-else
                :data="build.clickDetail ? build.clickDetail.builds || [] : []"
                :cell-style="cellStyle"
                style="width: 100%">
                <el-table-column
                  prop="workspace_name"
                  label="流水线空间"
                  width="230">
                </el-table-column>
                <el-table-column
                  prop="pipeline_name"
                  label="流水线"
                  width="150">
                </el-table-column>
                <el-table-column
                  prop="when"
                  label="构建号"
                  width="140">
                  <template slot-scope="scope">
                    {{ scope.row.build_release_version ? scope.row.build_release_version : '#' + scope.row.build_number }} ({{scope.row.build_operator}})
                  </template>
                </el-table-column>
                <el-table-column
                  prop="code_commit"
                  label="CommitId"
                  width="160">
                  <template slot-scope="scope">
                    {{ scope.row.code_commit.substr(0, 10) }} ({{ scope.row.code_author }})
                  </template>
                </el-table-column>
                <el-table-column
                  prop="code_comment"
                  label="Comment"
                  width="">
                </el-table-column>
              </el-table>
              <!-- <el-button style="margin-top: 8px; border-radius: 0px; padding: 5px 15px;" type="primary" size="mini">重新构建</el-button> -->
            </template>
            <template v-if="build.clickDetail && build.clickDetail.type == 'stage'">
              <div style="font-size: 14px; padding: 4px 3px 8px">
                阶段：{{ build.clickDetail ? build.clickDetail.stage ? build.clickDetail.stage.name : '' : '' }}
              </div>
              <el-table
                :data="build.clickDetail ? build.clickDetail.stage ? build.clickDetail.stage.jobs : [] : []"
                :cell-style="cellStyle"
                style="width: 100%">
                <el-table-column
                  prop="name"
                  label="任务"
                  width="180">
                </el-table-column>
                <el-table-column
                  prop="status"
                  label="状态"
                  width="90">
                  <template slot-scope="scope">
                    <span>
                      {{ jobStatusMap[scope.row.status] }}
                    </span>
                  </template>
                </el-table-column>
                <el-table-column
                  prop="result"
                  label="执行结果"
                  width="">
                  <template slot-scope="scope">
                    <span>
                      {{ scope.row.result }}
                    </span>
                  </template>
                </el-table-column>
              </el-table>
              <!-- <el-button style="margin-top: 8px; border-radius: 0px; padding: 5px 15px;" type="primary" size="mini">重新构建</el-button> -->
              <el-button style="margin-top: 8px; border-radius: 0px; padding: 5px 15px;" type="danger" 
                size="mini" v-if="(build.clickDetail.stage.status == 'doing')" @click="stageCancel(build.clickDetail.stage)">取消执行</el-button>
              <el-button style="margin-top: 8px; border-radius: 0px; padding: 5px 15px;" type="primary" 
                size="mini" v-if="(build.clickDetail.stage.status == 'canceled')" @click="stageReexec(build.clickDetail.stage)">重新执行</el-button>
            </template>
            <el-button style="margin-top: 8px; border-radius: 0px; padding: 5px 15px;" type="primary" 
                size="mini" v-if="(build.pipeline_run.status == 'error')" @click="stageRetry(build)">失败重试</el-button>
          </div>
        </div>
      </div>
      <div v-if="loadMore" style="text-align: center; margin: 15px 15px 5px;">
          <el-button style="border-radius:0px;" type="primary" @click="fetchBuilds()" size="small">加 载 更 多</el-button>
      </div>

      <div v-if="builds && builds.length == 0" style="text-align: center; margin-top: 30px; margin-left: -100px; color: #606266; font-size: 14px;">
        暂无流水线构建记录，<span @click="openBuildParams" class="build-span" style="color:#409EFF">执行流水线</span>
      </div>
    </div>

    <el-dialog title="执行流水线" :visible.sync="dialogVisible" :destroy-on-close="true" 
      :close-on-click-modal="false">
      <div v-loading="dialogLoading">
        <div class="dialogContent" style="padding: 0px 40px;">
          <el-form :model="buildParams" ref="" label-position="left" label-width="105px">
            <el-form-item label="流水线" prop="" :required="true">
              <el-input :disabled="true" style="width: 100%;" placeholder="" v-model="pipelineName" autocomplete="off" size="small"></el-input>
            </el-form-item>
            <el-form-item v-if="pipeline.workspace && pipeline.workspace.type == 'code'" label="构建分支" prop="" :required="true">
              <el-select v-model="buildParams.branch" placeholder="请选择构建分支" size="small" style="width: 100%" 
                :loading="branchLoading" allow-create filterable>
                <el-option v-for="b in branches" :key="b.name" :label="b.name" :value="b.name"></el-option>
              </el-select>
              <!-- <el-input v-else style="width: 100%;" placeholder="请输入构建分支" v-model="buildParams.branch" autocomplete="off" size="small"></el-input> -->
            </el-form-item>
            <el-form-item v-if="pipeline.workspace && pipeline.workspace.type == 'custom'" label="" prop="" label-width="0px" :required="true">
              <el-row style="margin-bottom: 5px; margin-top: 8px;">
                <el-col :span="10" style="background-color: #F5F7FA; padding-left: 10px;">
                  <div class="border-span-header">
                    流水线空间
                  </div>
                </el-col>
                <el-col :span="5" style="background-color: #F5F7FA">
                  <div class="border-span-header">
                    流水线
                  </div>
                </el-col>
                <el-col :span="9" style="background-color: #F5F7FA">
                  <div class="border-span-header">
                    构建号
                  </div>
                </el-col>
                <!-- <el-col :span="5"><div style="width: 100px;"></div></el-col> -->
              </el-row>
              <el-row style="padding-bottom: 0px;" v-for="(d, i) in pipeline.pipeline ? pipeline.pipeline.triggers : []" :key="i">
                <el-col :span="10">
                  <div class="border-span-header" style="margin-right: 10px;">
                    {{ d.workspace_name }}
                  </div>
                </el-col>
                <el-col :span="5">
                  <div class="border-span-header">
                    {{ d.pipeline_name }}
                  </div>
                </el-col>
                <el-col :span="9">
                  <div class="border-span-header">
                    <el-select v-model="buildParams[d.pipeline].buildId" placeholder="流水线空间" size="small" style="width: 100%;"
                      >
                      <el-option v-for="w in pipelineBuilds[d.pipeline] || []" :key="w.pipeline_run.id" 
                        :label="buildPipilineLabel(w)" :value="w.pipeline_run.id"></el-option>
                    </el-select>
                  </div>
                </el-col>
              </el-row>
            </el-form-item>
          </el-form>
        </div>
        <div slot="footer" class="dialogFooter" style="margin-top: 20px;">
          <el-button @click="dialogVisible = false" style="margin-right: 20px;" >取 消</el-button>
          <el-button type="primary" @click="buildPipeline">确 定</el-button>
        </div>
      </div>
    </el-dialog>

    <el-dialog :title="'执行阶段-' + manualStage.stage.name" :visible.sync="manualDialogVisible" :destroy-on-close="true" 
      @close="manualStage = {stage: {}, params: {}};" :close-on-click-modal="false">
      <div v-loading="dialogLoading">
        <div class="dialogContent" style="padding: 0px 40px;">
          <el-form :model="manualStage" ref="" label-position="left" label-width="105px">
            <el-form-item label="阶段参数" prop="">
              <el-row style="margin-bottom: 5px; margin-top: 2px;">
                <el-col :span="11" style="background-color: #F5F7FA; padding-left: 10px;">
                  <div class="border-span-header">参数</div>
                </el-col>
                <el-col :span="13" style="background-color: #F5F7FA">
                  <div class="border-span-header">参数值</div>
                </el-col>
              </el-row>
              <el-row style="padding-bottom: 5px;" v-for="(d, i) in manualStage.custom_params" :key="i">
                <el-col :span="11">
                  <div class="border-span-header">
                    <el-input v-model="d.param" disabled size="small" style="padding-right: 10px" placeholder="阶段参数"></el-input>
                  </div>
                </el-col>
                <el-col :span="13">
                  <div class="border-span-header">
                    <el-input v-model="d.value" size="small" placeholder="请输入参数默认值"></el-input>
                  </div>
                </el-col>
              </el-row>
            </el-form-item>
          </el-form>
          <template v-for="(job, i) in manualStage.stage.jobs || []">
            <div :key="i" v-if="manualJobComponent[job.plugin_key]">
              <component v-if="manualStage.job_params[job.plugin_key]" v-bind:is="manualJobComponent[job.plugin_key]" :params="manualStage.job_params[job.plugin_key]"></component>
            </div>
          </template>
        </div>
        <div slot="footer" class="dialogFooter" style="margin-top: 20px;">
          <el-button @click="manualDialogVisible = false" style="margin-right: 20px;" >取 消</el-button>
          <el-button type="primary" @click="manualExec">执 行</el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { StatusIcon } from '@/views/pipeline/components'
import { getPipeline, listRepoBranches } from '@/api/pipeline/pipeline'
import { listBuilds, buildPipeline, manualExec, stageRetry, stageCancel, stageReexec } from '@/api/pipeline/build'
import { manualCheck, Release } from '@/views/pipeline/plugin-manual'
import { Message } from 'element-ui'
import { del } from 'vue'

export default {
  name: 'PipelineWorkspace',
  components: {
    Clusterbar,
    Release,
    StatusIcon,
  },
  sse: {cleanup: true},
  data() {
    return {
      titleName: [],
      search_name: '',
      users: [],
      cellStyle: {border: 0, padding: '1px 0', 'line-height': '35px'},
      maxHeight: window.innerHeight - this.$contentHeight,
      loading: true,
      dialogVisible: false,
      pipeline: {},
      pipelineName: '',
      builds: null,
      buildDetails: {},
      buildParams: {},
      pipelineSSE: null,
      refreshExecTimer: 0,
      refreshStages: {},
      loadMore: false,
      dialogLoading: false,
      manualDialogVisible: false,
      manualStage: {
        stage: {},
        job_params: {},
        custom_params: []
      },
      manualJobComponent: {
        release: Release
      },
      pipelineBuilds: {},
      branches: [],
      branchLoading: false,
      jobStatusMap: {
        'ok': '执行成功',
        'error': '执行失败',
        'wait': '未执行',
        'doing': '执行中',
        'cancel': '取消',
        'canceled': '已取消',
      }
    }
  },
  created() {
    this.loading = true
    this.fetchPipeline();
    this.fetchBuilds();
    this.fetchPipelineSSE();
  },
  beforeDestroy() {
    if(this.refreshExecTimer) {
      clearTimeout(this.refreshExecTimer)
    }
  },
  mounted() {
    const that = this
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - this.$contentHeight
        that.maxHeight = heightStyle
      })()
    }
  },
  computed: {
    pipelineId() {
      return this.$route.params.pipelineId
    },
  },
  methods: {
    fetchPipeline() {
      getPipeline(this.pipelineId).then((response) => {
        this.pipeline = response.data || {};
        if (this.pipeline){
          this.titleName = ["流水线", this.pipeline.pipeline.name]
          this.pipelineName = this.pipeline.pipeline.name
          if(this.pipeline.workspace.type == 'custom') {
            for(let t of this.pipeline.pipeline.triggers || []) {
              // this.buildParams[t.pipeline] = {}
              this.$set(this.buildParams, t.pipeline, t)
            }
          }
        }
      }).catch(() => {
        
      })
    },
    fetchBuilds(lastBuildNumber) {
      this.loading = true
      if(lastBuildNumber == undefined) {
        lastBuildNumber = 0
        if(this.builds && this.builds.length > 0) {
          lastBuildNumber = this.builds[this.builds.length - 1].pipeline_run.build_number
        }
      }
      listBuilds(this.pipelineId, lastBuildNumber).then((response) => {
        this.loading = false
        let res = response.data || []
        if(lastBuildNumber == 0) {
          this.$set(this, 'builds', res)
        } else {
          for(let r of res) this.builds.push(r)
        }
        if(res.length == 20) this.loadMore = true
        else this.loadMore = false
        this.processExecTime()
      }).catch(() => {
        this.loading = false
      })
    },
    processExecTime() {
      var hasDoing = false
      for(let build of this.builds) {
        for(let s of build.stages_run) {
          if(build.pipeline_run.status == 'doing') {
            if(s.status == 'doing') {
              if(!this.refreshStages[s.id]) {
                var endTime = new Date();
                var execTime = new Date(s.exec_time);
                var diffTime = Math.floor((endTime.getTime()-execTime.getTime()) / 1000)
                if (diffTime <= 0) diffTime = 0
                this.$set(this.refreshStages, s.id, diffTime)
              }
              hasDoing = true
            } else if(this.refreshStages[s.id]) {
              // delete this.refreshStages[s.id]
              this.$delete(this.refreshStages, s.id)
            }
          } else if(this.refreshStages[s.id]) {
            this.$delete(this.refreshStages, s.id)
          }
        }
      }
      if(!this.refreshExecTimer  && hasDoing) {
        this.refreshExecTime()
      } else if(!hasDoing && this.refreshExecTimer) {
        clearTimeout(this.refreshExecTimer)
        this.refreshExecTimer = 0
      }
    },
    refreshExecTime() {
      let that = this
      if(this.refreshExecTimer) {
        clearTimeout(this.refreshExecTimer)
        this.refreshExecTimer = 0
      }
      this.refreshExecTimer = setTimeout(function () {
        let rs = that.refreshStages
        for(let s in rs) {
          let t = rs[s]
          if(t != undefined) that.$set(that.refreshStages, s, t + 1)
        }
        // that.$set(that, 'refresthStages', rs)
        that.refreshExecTime()
      }, 1000);
    },
    fetchPipelineSSE() {
      let url = `/api/v1/pipeline/pipeline/${this.pipelineId}/sse`
      this.pipelineSSE = this.$sse.create({
        url: url,
        withCredentials: false,
        format: 'plain',
        polyfill: true,
      });
      this.pipelineSSE.on("message", (res) => {
        // console.log(res)
        if(res && res != "\n") {
          let data = JSON.parse(res)
          if(data.pipeline_run) {
            for(let i in this.builds){
              let build = this.builds[i]
              if(build.pipeline_run.id == data.pipeline_run.id) {
                if (build.clickDetail) {
                  let clickDetail = build.clickDetail
                  if(clickDetail.type == 'stage') {
                    for(let s of data.stages_run) {
                      if (s.id == clickDetail.stage.id) {
                        clickDetail.stage = s
                        break
                      }
                    }
                  }
                  data.clickDetail = clickDetail
                }
                this.$set(this.builds, i, data)
                this.processExecTime()
                break
              }
            }
          }
        }
      })
      this.pipelineSSE.connect().then(() => {
        console.log('[info] connected', 'system')

        // this.pipelineSSE.disconnect()
      }).catch(() => {
        console.log('[error] failed to connect', 'system')
      })
      this.pipelineSSE.on('error', () => { // eslint-disable-line
        console.log('[error] disconnected, automatically re-attempting connection', 'system')
      })
    },
    releaseVersion(stage) {
      for(let s of stage.jobs) {
        if(s.plugin_key == 'release' && s.params.version && s.status != 'wait') {
          return ' - ' + s.params.version
        }
      }
    },
    nameClick: function(id) {
      this.$router.push({name: "pipelineBuilds", params: { pipelineId: id },});
    },
    nameSearch: function(val) {
      this.search_name = val
    },
    openBuildParams() {
      this.dialogVisible = true
      // this.dialogLoading = true
      if(this.pipeline.workspace && this.pipeline.workspace.type == 'custom') {
        this.pipelineBuilds = {}
        for(let t of this.pipeline.pipeline.triggers || []) {
          // 获取触发源的每条流水线最新50条构建
          this.fetchPipelineBuilds(t.pipeline)
        }
      }else if (this.pipeline.workspace && this.pipeline.workspace.code){
          this.branchLoading = true
          // 获取github/gitlab/gitee代码仓库分支
          listRepoBranches(this.pipelineId).then((response) => {
            this.branchLoading = false
            this.branches = response.data || []
          }).catch(() => {
            this.branchLoading = false
          })
        
      }
    },
    fetchPipelineBuilds(pipelineId) {
      listBuilds(pipelineId, 0, "ok", 50).then((response) => {
        let res = response.data || []
        this.$set(this.pipelineBuilds, pipelineId, res)
        if(res.length) {
          // this.buildParams[pipelineId].build = {buildId: res[0].pipeline_run.id, pipelineId: res[0].pipeline_run.pipeline_id}
          this.$set(this.buildParams[pipelineId], "buildId", res[0].pipeline_run.id)
        }
        // this.pipelineBuilds[pipelineId] = res
      }).catch(() => {
        this.loading = false
      })
    },
    buildPipeline: function() {
      if(!this.pipeline.workspace) {
        Message.error("获取流水线空间参数失败，请刷新重试")
        return
      }
      let params = {}
      if(this.pipeline.workspace.type == 'code') {
        if(!this.buildParams.branch) {
          Message.error("请输入构建分支")
          return
        }
        params = {
          branch: this.buildParams.branch
        }
      } else {
        let build_ids = []
        for(let pipelineId in this.buildParams) {
          let build = this.buildParams[pipelineId]
          let buildId = build.buildId
          for(let b of this.pipelineBuilds[pipelineId]) {
            if(buildId == b.pipeline_run.id) {
              let info = {
                workspace_id: build.workspace,
                workspace_name: build.workspace_name,
                pipeline_id: parseInt(pipelineId),
                pipeline_name: build.pipeline_name,
                build_id: this.buildParams[pipelineId].buildId,
                build_number: b.pipeline_run.build_number,
                build_operator: b.pipeline_run.operator,
                build_release_version: b.stages_run[b.stages_run.length - 1].env.RELEASE_VERSION || "",
                code_branch: b.pipeline_run.env["PIPELINE_CODE_BRANCH"],
                code_author: b.pipeline_run.env["PIPELINE_CODE_COMMIT_AUTHOR"],
                code_comment: b.pipeline_run.env["PIPELINE_CODE_COMMIT_MESSAGE"],
                code_commit: b.pipeline_run.env["PIPELINE_CODE_COMMIT_ID"],
                code_commit_time: b.pipeline_run.env["PIPELINE_CODE_COMMIT_TIME"],
                is_build: true
              }
              build_ids.push(info)
              break
            }
          }
        }
        params.build_ids = build_ids
      }
      this.dialogLoading = true
      buildPipeline(this.pipelineId, params).then((response) => {
        this.$message({message: '构建成功', type: 'success'});
        this.dialogLoading = false
        this.fetchBuilds(0)
        this.dialogVisible = false
      }).catch( (err) => {
        this.dialogLoading = false
      })
    },
    getStageStatus(status) {
      if(status == 'ok') return 'success'
      if(status == 'error') return 'error'
      if(status == 'wait') return 'wait'
      if(status == 'doing') return 'process'
      if(status == 'cancel') return 'process'
      if(status == 'canceled') return 'process'
      if(status == 'pause') return 'process'
    },
    getBuildStatusColor(status) {
      if(status == 'ok') return '#67c23a'
      if(status == 'error') return '#DC143C'
      if(status == 'doing') return '#E6A23C'
      return ''
    },
    getStageExecTime(execTimeStr, endTimeStr) {
      if(!endTimeStr) var endTime = new Date();
      else var endTime = new Date(endTimeStr)

      var execTime = new Date(execTimeStr)
      var diffTime = Math.floor((endTime.getTime()-execTime.getTime()) / 1000)
      if (diffTime < 0) diffTime = 0
      return this.getStageExecTimeStr(diffTime)
    },
    
    getStageExecTimeStr(diffTime) {
      var stageTime = ''

      var days = Math.floor(diffTime / (24*3600))
      if (days) stageTime = days + 'd'

      var leave1 = diffTime % (24*3600)
      var hours = Math.floor(leave1/(3600))
      if(hours) stageTime += hours + 'h'

      var leave2=leave1%(3600)        //计算小时数后剩余的毫秒数
      var minutes=Math.floor(leave2/(60))
      if(minutes) stageTime += minutes + 'm'

      var leave3 = leave2 % (60)
      var seconds=Math.round(leave3)
      if(seconds) stageTime += seconds + 's'
      if(!stageTime) stageTime = '1s'
      return stageTime
    },
    
    clickBuildDetail(build, type, stage) {
      var clickDetail = build.clickDetail
      if(clickDetail){
        if(type == 'source' && clickDetail.type == type) {
          this.$set(build, 'clickDetail', undefined)
          return
        }
        if(type == 'stage' && clickDetail.stage && clickDetail.stage.id == stage.id) {
          this.$set(build, 'clickDetail', undefined)
          return
        }
      }
      if(type == 'source') {
        if(this.pipeline.workspace.type == 'code') {
          let commit = {
            commitId: build.pipeline_run.env.PIPELINE_CODE_COMMIT_ID,
            author: build.pipeline_run.env.PIPELINE_CODE_COMMIT_AUTHOR,
            message: build.pipeline_run.env.PIPELINE_CODE_COMMIT_MESSAGE,
            when: build.pipeline_run.env.PIPELINE_CODE_COMMIT_TIME,
          }
          this.$set(build, 'clickDetail', {type: type, commit: [commit]})
        } else {
          this.$set(build, 'clickDetail', {type: type, builds: build.pipeline_run.params.build_ids})
        }
      } else {
        this.$set(build, 'clickDetail', {type: type, stage: stage})
      }
    },
    clickBuildNumber(build) {
      this.$router.push({name: 'pipelineBuildDetail', params: {buildId: build.pipeline_run.id}})
    },
    openManualStageDialog(stage) {
      let custom_params = []
      if(stage.custom_params) {
        for(let k in stage.custom_params) {
          custom_params.push({param: k, value: stage.custom_params[k]})
        }
      }
      let job_params = {}
      for(let j of stage.jobs) {
        job_params[j.plugin_key] = j.params || {}
      }
      this.$set(this.manualStage, 'job_params', job_params)
      this.$set(this.manualStage, 'custom_params', custom_params)
      this.$set(this.manualStage, 'stage', stage)
      this.manualDialogVisible = true
    },
    async manualExec() {
      let parmas = {
        stage_run_id: this.manualStage.stage.id,
        job_params: this.manualStage.job_params
      }
      let custom_params = {}
      if(this.manualStage.custom_params) {
        for(let p of this.manualStage.custom_params) {
          custom_params[p.param] = p.value
        }
      }
      parmas['custom_params'] = custom_params
      if(!parmas.stage_run_id) {
        Message.error("获取执行阶段id错误，请刷新重试")
        return
      }
      this.dialogLoading = true
      let err = await manualCheck(this.pipeline, this.manualStage.stage, this.manualStage.job_params)
      if (err) {
        Message.error(err)
        this.dialogLoading = false
        return
      }
      manualExec(parmas).then((response) => {
        this.$message({message: '下发任务成功', type: 'success'});
        this.dialogLoading = false
        // this.fetchBuilds(0)
        this.manualDialogVisible = false
      }).catch( (err) => {
        this.dialogLoading = false
      })
    },
    stageRetry(build) {
      if(!build) {
        Message.error("获取构建信息失败，请刷新重试")
        return
      }
      if(!build.stages_run) {
        Message.error("获取构建阶段失败，请刷新重试")
        return
      }
      let stageId = 0
      for(let s of build.stages_run) {
        if(s.status == 'error') {
          stageId = s.id
        }
      }
      if(!stageId) {
        Message.error("未获取到执行失败的阶段，请刷新重试")
        return
      }
      this.loading = true
      let parmas = {stage_run_id: stageId}
      stageRetry(parmas).then((response) => {
        this.$message({message: '重试成功', type: 'success'});
        // this.fetchBuilds(0)
        this.loading = false
      }).catch( (err) => {
        this.loading = false
      })
    },
    stageReexec(stage) {
      if(!stage) {
        Message.error("获取构建阶段失败，请刷新重试")
        return
      }
      if(!stage.id) {
        Message.error("获取构建阶段失败，请刷新重试")
        return
      }
      this.loading = true
      let parmas = {stage_run_id: stage.id}
      stageReexec(parmas).then((response) => {
        this.$message({message: '重新执行成功', type: 'success'});
        // this.fetchBuilds(0)
        this.loading = false
      }).catch( (err) => {
        this.loading = false
      })
    },
    stageCancel(stage) {
      if(!stage) {
        Message.error("获取构建阶段失败，请刷新重试")
        return
      }
      if(!stage.id) {
        Message.error("获取构建阶段失败，请刷新重试")
        return
      }
      this.loading = true
      let parmas = {stage_run_id: stage.id}
      stageCancel(parmas).then((response) => {
        this.$message({message: '取消成功', type: 'success'});
        // this.fetchBuilds(0)
        this.loading = false
      }).catch( (err) => {
        this.loading = false
      })
    },
    buildPipilineLabel(build) {
      let build_num = '#' + build.pipeline_run.build_number
      let lastStage = build.stages_run[build.stages_run.length - 1]
      if(lastStage && lastStage.env.RELEASE_VERSION) {
        build_num = lastStage.env.RELEASE_VERSION
      }
      return `${build_num} (${build.pipeline_run.operator})`
    },
  },
}
</script>

<style lang="scss" scoped>

.table-fix {
  height: calc(100% - 100px);
}

.name-class {
  cursor: pointer;
}
.name-class:hover {
  color: #409EFF;
}

.dashboard-container-build {
  margin: 10px 20px;
  overflow:scroll;
}

.build-list {
  width: 100%;
  padding: 0px 10px;
  // border: 1px solid #EBEEF5;
  color: #606266;
  // margin-top: 10px;
}
// .build-list:hover {
//   border: 1px solid #EBEEF5;
// }
.build-info {
  display: flex;
  position: relative;
  box-sizing: border-box;
}
.build-info__left {
  display: flex;
  float: left;
  margin-right: 10px;
  padding: 15px 0;
}
.build-info__left:hover {
  cursor: pointer;
}
.build-info__left-number {
  display: inline-block;
  float: left;
  width: 80px;
  margin-right: 10px;
  font-size: 13px;
}
.build-info__left-number-inner {
  font-size: 17px;
  height: 22px;
  line-height: 22px;
  font-weight: 500;
}
.build-info__left__number:hover {
  cursor: pointer;
  color: #409EFF;
}
.build-info__left-branch {
  font-size: 13px;
  height: 22px;
  line-height: 22px;
  vertical-align: middle;
}
.build-info__left-op {
  display: inline-block;
  float: left;
  width: 130px;
  font-size: 13px;
  margin-top: 1px;
}
.el-steps {
  margin-top: 9px;
  padding: 6px 8%;
}
.el-steps-title {
  font-size: 14px;
  font-weight: 400;
  width: 120px;
}
.el-steps-title-name:hover{
  cursor: pointer;
  // font-size: 15px;
  // font-weight: 450;
  color: #409EFF;
}
.el-steps-stage-exectime {
  display: inline-flex;
  vertical-align: 1px;
  margin-left: 6px;
  font-size: 14px;
  font-weight: 400;
}
.build-stage {
}

.build-span{

}
.build-span:hover{
  cursor: pointer;
}

.refresh-rotate {
  animation: loading-rotate 1.5s cubic-bezier(0.29, 0.99, 0.73, 0.02) infinite;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
</style>
<style lang="scss">
.dashboard-container-build {

  .build-list .el-steps--simple {
    border-radius: 0px;
  }
  .build-list .el-step.is-simple .el-step__head {
    display: none;
  }
  .build-detail .el-table::before {
    height: 0px;
  }
  .build-detail .el-table .cell {
    font-size: 13px;
    font-weight: 400;
  }
  .build-detail .el-table td, .build-detail .el-table th {
    padding: 0px 0;
  }
}
</style>
