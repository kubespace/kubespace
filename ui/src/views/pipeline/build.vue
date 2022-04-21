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
                <div class="build-info__left-number-inner" :style="{color: getBuildStatusColor(build.pipeline_run.status)}">
                  <template v-if="build.pipeline_run.status == 'ok'">
                    <i class="el-icon-success"></i>
                  </template>
                  <template v-if="build.pipeline_run.status == 'error'">
                    <i class="el-icon-error"></i>
                  </template>
                  <template v-if="build.pipeline_run.status == 'wait'">
                    <i class="el-icon-remove"></i>
                  </template>
                  <template v-if="build.pipeline_run.status == 'doing'">
                    <i class="el-icon-refresh refresh-rotate"></i>
                  </template>
                  <template v-if="build.pipeline_run.status == 'pause'">
                    <!-- <i class="el-icon-video-pause"></i> -->
                    <svg-icon icon-class="pause" />
                  </template>
                  <span class="build-info__left__number" @click="clickBuildNumber(build)"> #{{ build.pipeline_run.build_number }}</span>
                  <!-- #{{ build.pipeline_run.build_number }} -->
                </div>
                <div class="build-info__left-branch">
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
                    <div style="margin-left: 1px;"  @click="clickBuildDetail(build, 'stage', stage)">{{ stage.name }}{{ releaseVersion(stage) }}</div>
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
                      <template v-if="stage.status == 'pause'">
                        <!-- <svg-icon icon-class="pause" style="font-size: 18px;"/>
                        <div class="el-steps-stage-exectime">
                          --
                        </div> -->
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
              <el-table
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
                  width="200">
                </el-table-column>
                <el-table-column
                  prop="result"
                  label="返回"
                  width="">
                  <template slot-scope="scope">
                    <span>
                      {{ scope.row.result }}
                    </span>
                  </template>
                </el-table-column>
              </el-table>
              <!-- <el-button style="margin-top: 8px; border-radius: 0px; padding: 5px 15px;" type="primary" size="mini">重新构建</el-button> -->
              
            </template>
            <el-button style="margin-top: 8px; border-radius: 0px; padding: 5px 15px;" type="primary" 
                size="mini" :disabled="!(build.pipeline_run.status == 'error')" @click="stageRetry(build)">失败重试</el-button>
          </div>
        </div>
      </div>
      <div v-if="loadMore" style="text-align: center; margin: 15px 15px 5px;">
          <el-button style="border-radius:0px;" type="primary" @click="fetchBuilds" size="small">加 载 更 多</el-button>
      </div>

      <div v-if="builds && builds.length == 0" style="text-align: center; margin-top: 30px; margin-left: -100px; color: #606266; font-size: 14px;">
        暂无流水线构建记录，<span @click="openBuildParams" class="build-span" style="color:#409EFF">执行流水线</span>
      </div>
    </div>

    <el-dialog title="执行流水线" :visible.sync="dialogVisible" :destroy-on-close="true" 
      @close="buildParams = {};" :close-on-click-modal="false">
      <div v-loading="dialogLoading">
        <div class="dialogContent" style="padding: 0px 40px;">
          <el-form :model="buildParams" ref="" label-position="left" label-width="105px">
            <el-form-item label="流水线" prop="" :required="true">
              <el-input :disabled="true" style="width: 100%;" placeholder="" v-model="pipelineName" autocomplete="off" size="small"></el-input>
            </el-form-item>
            <el-form-item label="构建分支" prop="" :required="true">
              <el-input style="width: 100%;" placeholder="请输入构建分支" v-model="buildParams.branch" autocomplete="off" size="small"></el-input>
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
import { getPipeline } from '@/api/pipeline/pipeline'
import { listBuilds, buildPipeline, manualExec, stageRetry } from '@/api/pipeline/build'
import { manualCheck, Release } from '@/views/pipeline/plugin-manual'
import { Message } from 'element-ui'

export default {
  name: 'PipelineWorkspace',
  components: {
    Clusterbar,
    Release
  },
  data() {
    return {
      titleName: [],
      search_name: '',
      users: [],
      cellStyle: {border: 0, padding: '1px 0', 'line-height': '35px'},
      maxHeight: window.innerHeight - 145,
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
    this.pipelineSSE.disconnect()
    if(this.refreshExecTimer) {
      clearTimeout(this.refreshExecTimer)
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
                this.$set(this.refreshStages, s.id, Math.floor((endTime.getTime()-execTime.getTime()) / 1000))
                // this.refreshStages[s.id] = 
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
    // fetchPipelineSSE() {
    //   let url = `/api/v1/pipeline/pipeline/${this.pipelineId}/sse`
    //   this.pipelineSSE = new EventSource(url);
    //   this.pipelineSSE.addEventListener('message', event => {
    //     // console.log(event.data);
    //     if(event.data && event.data != "\n") {
    //       let data = JSON.parse(event.data)
    //       // console.log(data)
    //       if(data.object) {
    //         let obj = data.object
    //         for(let i in this.builds){
    //           let build = this.builds[i]
    //           if(build.pipeline_run.id == obj.pipeline_run.id) {
    //             this.$set(this.builds, i, obj)
    //             this.processExecTime()
    //             break
    //           }
    //         }
    //       }
    //     }
    //   });
    //   this.pipelineSSE.addEventListener('error', event => {
    //     if (event.readyState == EventSource.CLOSED) {
    //       console.log('event was closed');
    //     };
    //     console.log(event)
    //   });
    //   this.pipelineSSE.addEventListener('close', event => {
    //     console.log(event.type);
    //     this.pipelineSSE.close();
    //   });
    // },
    fetchPipelineSSE() {
      let url = `/api/v1/pipeline/pipeline/${this.pipelineId}/sse`
      this.pipelineSSE = this.$sse.create({
        url: url,
        includeCredentials: false,
        format: 'plain'
      });
      this.pipelineSSE.on("message", (res) => {
        console.log(res)
        if(res && res != "\n") {
          let data = JSON.parse(res)
          // console.log(data)
          if(data.object) {
            let obj = data.object
            for(let i in this.builds){
              let build = this.builds[i]
              if(build.pipeline_run.id == obj.pipeline_run.id) {
                this.$set(this.builds, i, obj)
                this.processExecTime()
                break
              }
            }
          }
        }
      })
      this.pipelineSSE.connect().then(() => {
        console.log('[info] connected', 'system')
      }).catch(() => {
        console.log('[error] failed to connect', 'system')
      })
      this.pipelineSSE.on('error', () => { // eslint-disable-line
        console.log('[error] disconnected, automatically re-attempting connection', 'system')
      })
    },
    releaseVersion(stage) {
      for(let s of stage.jobs) {
        if(s.plugin_key == 'release' && s.params.version) {
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
    },
    buildPipeline: function() {
      if(!this.buildParams.branch) {
        Message.error("请输入构建分支")
        return
      }
      this.dialogLoading = true
      buildPipeline(this.pipelineId, this.buildParams).then((response) => {
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
      if(status == 'cancel') return 'wait'
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
        let commit = {
          commitId: build.pipeline_run.env.PIPELINE_CODE_COMMIT_ID,
          author: build.pipeline_run.env.PIPELINE_CODE_COMMIT_AUTHOR,
          message: build.pipeline_run.env.PIPELINE_CODE_COMMIT_MESSAGE,
          when: build.pipeline_run.env.PIPELINE_CODE_COMMIT_TIME,
        }
        this.$set(build, 'clickDetail', {type: type, commit: [commit]})
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
      console.log(this.manualStage)
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
        Message.error("获取构建阶段失败，请刷新重试")
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
    }
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
.el-steps-title:hover{
  cursor: pointer;
  font-size: 15px;
  font-weight: 450;
  // color: #81bd63;
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

@-webkit-keyframes rotation{
    from {-webkit-transform: rotate(0deg);}
    to {-webkit-transform: rotate(360deg);}
}

.refresh-rotate {
  -webkit-transform: rotate(360deg);
  animation: rotation 2s linear infinite;
  -moz-animation: rotation 2s linear infinite;
  -webkit-animation: rotation 2s linear infinite;
  -o-animation: rotation 2s linear infinite;
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
