<template>
  <div>
    <clusterbar :titleName="titleName" :titleLink="['pipeline', 'pipelineBuilds']"/>
    <div v-loading="loading" class="dashboard-container detail-dashboard dashboard-container-build-detail" style="margin-left: 20px; margin-right: 20px;" :style="{height: maxHeight + 'px', }" :max-height="maxHeight">
      <div style="padding: 5px 0px 0px;">
        <!-- <div>构建信息</div> -->
        <el-form label-position="left" inline class="pod-item" label-width="80px" 
          style="margin: 3px 0px 10px 0px; border: 1px solid #EBEEF5; box-shadow: none; padding: 5px 20px;">
          <el-form-item label="构建号">
            <span :style="{color: statusColorMap[build.pipeline_run.status], 'font-size': '16px'}">
              <i :class="statusIconMap[build.pipeline_run.status]"></i>
              #{{ build.pipeline_run.build_number }}
            </span>
          </el-form-item>
          <el-form-item label="构建时间">
            <span>{{ $dateFormat(build.pipeline_run.create_time) }}</span>
          </el-form-item>
          <el-form-item label="CommitId">
            <span >{{ getBuildCommitId() }}</span>
          </el-form-item>
          <el-form-item label="构建人">
            <span style="">{{ build.pipeline_run.operator }}</span>
          </el-form-item>
          <el-form-item label="构建分支">
            <span>{{ build.pipeline_run.env ? build.pipeline_run.env['PIPELINE_CODE_BRANCH'] : '' }}</span>
          </el-form-item>
          <el-form-item label="Comment">
            <span >{{ getBuildCommitComment() }}</span>
          </el-form-item>
        </el-form>
      </div>
      <div>
        <!-- <div>阶段任务</div> -->
        <div label-position="left" inline class="pod-item" label-width="80px" :style="{height: maxHeight - 80 + 'px'}" 
          style="margin: 0px; border: 1px solid #EBEEF5; box-shadow: none; font-size: 14px; padding: 0px;">
          <el-container style="height: 100%;">
            <el-aside width="240px" hight="100%" 
              style="border-right: 1px solid #EBEEF5; height: 100%; padding: 10px 20px; line-height: 25px; color: #606266">
              <div :style="{color: statusColorMap[stage.status]}" v-for="stage in build.stages_run" :key="stage.id">
                <div @click="clickStage(stage)" class="click-main-content">
                  {{ stage.name }}
                </div>
                <div style="margin-left: 15px;" v-for="job in stage.jobs" :key="job.id" class="click-main-content"
                  @click="clickStage(stage, job)">
                  <i :class="statusIconMap[job.status]"></i> {{ job.name }}
                </div>
              </div>
            </el-aside>

            <el-main style="padding: 3px 0px" v-if="mainContent.type == 'stage'">
              <el-form label-position="left" inline class="pod-item" label-width="80px" 
                style="margin: 3px 0px 0px 0px; border: 0px solid #EBEEF5; box-shadow: none; padding: 5px 20px;">
                <el-form-item label="阶段">
                  <span :style="{color: statusColorMap[mainContent.mainStage.status], 'font-size': '14px'}">
                    <i :class="statusIconMap[mainContent.mainStage.status]"></i>
                    {{ mainContent.mainStage.name }}
                  </span>
                </el-form-item>
                <el-form-item label="执行时间">
                  <span>{{ $dateFormat(mainContent.mainStage.exec_time) }}</span>
                </el-form-item>
                <el-form-item label="完成时间" v-if="mainContent.mainStage.status == 'ok' || mainContent.mainStage.status=='error'">
                  <span >{{ $dateFormat(mainContent.mainStage.update_time) }}</span>
                </el-form-item>
              </el-form>
              <div style="margin-left: 20px; margin-right: 20px; color: #99a9bf" class="stage-params">
                
                <el-table
                  :data="mainStageEnvs"
                  class="table-fix"
                  tooltip-effect="dark"
                  :max-height="maxHeight-150"
                  style="width: 100%"
                  v-loading="loading"
                  :cell-style="cellStyle"
                  :default-sort = "{prop: 'name'}"
                  row-key="name"
                >
                  <el-table-column prop="name" label="流水线环境参数" show-overflow-tooltip>
                  </el-table-column>
                  <el-table-column prop="value" label="参数值">
                  </el-table-column>
                </el-table>
              </div>
            </el-main>

            <el-main style="padding: 3px 0px" v-if="mainContent.type == 'job'">
              <el-form label-position="left" inline class="pod-item" label-width="80px" 
                style="margin: 3px 0px 0px 0px; border: 0px solid #EBEEF5; box-shadow: none; padding: 5px 20px;">
                <el-form-item label="任务">
                  <span :style="{color: statusColorMap[mainContent.mainJob.status], 'font-size': '14px'}">
                    <i :class="statusIconMap[mainContent.mainJob.status]"></i>
                    {{ mainContent.mainJob.name }}
                  </span>
                </el-form-item>
                <el-form-item label="执行时间">
                  <span>{{ $dateFormat(mainContent.mainStage.exec_time) }}</span>
                </el-form-item>
                <el-form-item label="完成时间" v-if="mainContent.mainJob.status == 'ok' || mainContent.mainJob.status=='error'">
                  <span >{{ $dateFormat(mainContent.mainJob.update_time) }}</span>
                </el-form-item>
              </el-form>
              <div :style="{'height': maxHeight - 130 + 'px'}"
                style="overflow: scroll; color: #C0C4CC; margin: 0px 10px; padding: 10px 15px; line-height: 17px; 
                  white-space: pre-wrap; background-color: #303133; " id="jobLogDiv">
                <div>{{ mainContent.mainJob.status == 'wait' ? '待执行...' : mainContent.jobLog ? mainContent.jobLog : "" }}</div>
                <div v-if="mainContent.mainJob.status !='doing' && mainContent.mainJob.status !='wait'">执行结果：<br/>{{ mainContent.mainJob.result }}</div>
                <div v-if="mainContent.mainJob.status =='doing'"><i class="el-icon-loading"></i></div>
              </div>
            </el-main>
          </el-container>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { getBuild, getJobLog } from '@/api/pipeline/build'

export default {
  name: 'PipelineBuildDetail',
  components: {
    Clusterbar
  },
  data() {
    return {
      titleName: [],
      search_name: '',
      users: [],
      cellStyle: {border: 0, padding: '6px 0',},
      maxHeight: window.innerHeight - 130,
      loading: true,
      build: {pipeline_run: {}},
      buildSSE: null,
      statusIconMap: {
        ok: "el-icon-success",
        error: "el-icon-error",
        wait: "el-icon-remove",
        doing: "el-icon-refresh refresh-rotate",
      },
      statusColorMap: {
        ok: '#67c23a',
        error: '#DC143C',
        doing: '#E6A23C'
      },
      pipelines: [{name: "PIPELINE_CODE_COMMIT_ID", value: "https://github.com/openspacee/osp"}],
      mainContent: {
        type: "",
        mainStage: {},
        mainJob: {},
        jobLog: "",
      },
      jobLogSSE: null,
      jobLogId: '',
      runningStatus: ['doing', 'wait'],
      scrollToBottom: true,
    }
  },
  created() {
    this.getBuild();
  },
  destroyed() {
    if(this.buildSSE) this.buildSSE.close()
    if(this.jobLogSSE) this.jobLogSSE.close()
  },
  mounted() {
    const that = this
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - 130
        that.maxHeight = heightStyle
      })()
    }
  },
  computed: {
    buildId() {
      return this.$route.params.buildId
    },
    mainStageEnvs() {
      if(this.mainContent.mainStage.env) {
        let r = []
        for(let k in this.mainContent.mainStage.env) {
          r.push({name: k, value: this.mainContent.mainStage.env[k]})
        }
        return r
      }else {
        return []
      }
    }
  },
  methods: {
    getBuild() {
      this.loading = true
      getBuild(this.buildId).then((response) => {
        this.build = response.data;
        this.titleName = ["流水线", this.build.pipeline.name, '#' + this.build.pipeline_run.build_number]
        this.getMainContent()
        if(this.runningStatus.indexOf(this.build.pipeline_run.status) >= 0) {
          this.fetchBuildSSE()
        }
        this.loading = false
      }).catch(() => {
        this.loading = false
      })
    },
    getMainContent() {
      for(let s of this.build.stages_run) {
        if(s.status != 'ok') {
          for(let j of s.jobs) {
            if(j.status != 'ok') {
              this.mainContent = {
                type: 'job',
                mainStage: s,
                mainJob: j,
                jobLog: ''
              }
              this.getJobLog()
              return
            }
          }
        }
      }
      this.mainContent = {
        type: 'job',
        mainStage: this.build.stages_run[0],
        mainJob: this.build.mainStage.jobs[0],
        jobLog: ''
      }
      this.getJobLog()
    },
    fetchBuildSSE() {
      let url = `/api/v1/pipeline/build/${this.buildId}/sse`
      this.buildSSE = new EventSource(url);
      this.buildSSE.addEventListener('message', event => {
          // console.log(event.data);
          if(event.data) {
            let data = JSON.parse(event.data)
            // console.log(data)
            if(data.object) {
              let obj = data.object
              console.log(obj)
              if(obj.pipeline_run) {
                this.$set(this.build, 'pipeline_run', obj.pipeline_run)
              }
              if(obj.stages_run) {
                this.$set(this.build, 'stages_run', obj.stages_run)
              }
              if(this.runningStatus.indexOf(this.build.pipeline_run.status) == -1) {
                this.buildSSE.close()
              }
              for(let s of this.build.stages_run) {
                if(s.id == this.mainContent.mainStage.id) {
                  this.mainContent.mainStage = s
                  if(this.mainContent.mainJob.id) {
                    for(let j of s.jobs) {
                      if(j.id == this.mainContent.mainJob.id) {
                        this.mainContent.mainJob = j
                      }
                    }
                  }
                  break
                }
              }
              this.getJobLog()
            }
          }
      });
      this.buildSSE.addEventListener('error', event => {
        if (event.readyState == EventSource.CLOSED) {
          console.log('event was closed');
        };
      });
      this.buildSSE.addEventListener('close', event => {
        console.log(event.type);
        this.buildSSE.close();
      });
    },
    getBuildCommitId() {
      if(!this.build.pipeline_run) return ''
      if(!this.build.pipeline_run.env) return ''
      if(!this.build.pipeline_run.env['PIPELINE_CODE_COMMIT_ID']) return ''
      let id = this.build.pipeline_run.env['PIPELINE_CODE_COMMIT_ID'].substr(0, 10)
      let author = this.build.pipeline_run.env['PIPELINE_CODE_COMMIT_AUTHOR']
      return id + " (" + author + ")"
    },
    getBuildCommitComment() {
      if(!this.build.pipeline_run) return ''
      if(!this.build.pipeline_run.env) return ''
      if(!this.build.pipeline_run.env['PIPELINE_CODE_COMMIT_MESSAGE']) return ''
      return this.build.pipeline_run.env['PIPELINE_CODE_COMMIT_MESSAGE'].substr(0, 30)
    },
    clickStage(stage, job) {
      this.mainContent.type = job ? 'job': 'stage'
      this.mainContent.mainStage = stage
      this.mainContent.mainJob = job ? job : {}
      this.getJobLog()
    },
    getJobLog() {
      if(this.mainContent.type == 'job') {
        if(this.mainContent.mainJob.id != this.jobLogId) {
          this.mainContent.jobLog = ''
          this.jobLogId = this.mainContent.mainJob.id
          if(this.jobLogSSE) {
            this.jobLogSSE.close()
          }
          let withSSE = false
          if(this.runningStatus.indexOf(this.mainContent.mainJob.status) >= 0) {
            withSSE = true
          }
          
          if(withSSE) {
            this.fetchJobLogSSE(this.mainContent.mainJob.id)
          } else {
            getJobLog(this.mainContent.mainJob.id).then((response) => {
              this.$set(this.mainContent, 'jobLog', response.data)
            }).catch(() => {
            })
          }
        } else if(this.runningStatus.indexOf(this.mainContent.mainJob.status) == -1 && this.jobLogSSE) {
          this.jobLogSSE.close()
          getJobLog(this.mainContent.mainJob.id).then((response) => {
            this.$set(this.mainContent, 'jobLog', response.data)
            this.$nextTick(() => {
              if (that.scrollToBottom) {
                let logDiv = document.getElementById('jobLogDiv')
                logDiv.scrollTop = logDiv.scrollHeight // 滚动高度
              }
            })
          }).catch(() => {
          })
        }
      } else if(this.jobLogSSE) {
        this.jobLogSSE.close()
      }
    },
    fetchJobLogSSE(jobId) {
      let url = `/api/v1/pipeline/build/log/${jobId}/sse`
      this.jobLogSSE = new EventSource(url);
      this.jobLogSSE.addEventListener('message', event => {
        // console.log(event.data);
        if(event.data) {
          this.$set(this.mainContent, 'jobLog', event.data)
          let that = this
          this.$nextTick(() => {
            if (that.scrollToBottom) {
              let logDiv = document.getElementById('jobLogDiv')
              logDiv.scrollTop = logDiv.scrollHeight // 滚动高度
            }
          })
        }
      });
      this.jobLogSSE.addEventListener('error', event => {
        if (event.readyState == EventSource.CLOSED) {
          console.log('event was closed');
        };
      });
      this.jobLogSSE.addEventListener('close', event => {
        console.log(event.type);
        this.jobLogSSE.close();
      });
    }
  },
}
</script>

<style lang="scss" scoped>
.click-main-content:hover {
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
.dashboard-container-build-detail{
  .el-form-item__label{
    line-height: 26px;
  }
  .el-form-item__content {
    line-height: 26px;
  }

  .stage-params {
    .el-table::before {
      height: 0;
    }
  }
}
</style>
