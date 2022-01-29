<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :createFunc="buildPipeline" createDisplay="构建"/>
    <div class="dashboard-container dashboard-container-build" :style="{height: maxHeight + 'px'}" :max-height="maxHeight">
      <div class="build-list" v-for="build in builds" :key="build.pipeline_run.id">
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
                    <i class="el-icon-refresh"></i>
                  </template>
                  <span class="build-info__left__number" @click="clickBuildNumber(build)"> #{{ build.pipeline_run.build_number }}</span>
                  <!-- #{{ build.pipeline_run.build_number }} -->
                </div>
                <div class="build-info__left-branch">
                  <svg-icon icon-class="branch" /> develop
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
                  <div slot="title" class="el-steps-title" @click="clickBuildDetail(build, 'stage', stage)">
                    <div style="margin-left: 1px;">{{ stage.name }}</div>
                    <div style="margin-top: 3px;">
                      <template v-if="stage.status == 'ok'">
                        <i class="el-icon-circle-check" style="font-size: 18px;"></i>
                        <div class="el-steps-stage-exectime">
                          {{ getStageExecTime(stage.create_time, stage.update_time) }}
                        </div>
                      </template>
                      <template v-if="stage.status == 'error'">
                        <i class="el-icon-circle-close" style="font-size: 18px;"></i>
                        <div class="el-steps-stage-exectime">
                          {{ getStageExecTime(stage.create_time, stage.update_time) }}
                        </div>
                      </template>
                      <template v-if="stage.status == 'wait'">
                        <i class="el-icon-remove-outline" style="font-size: 18px;"></i>
                        <div class="el-steps-stage-exectime">
                          --
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
                :data="tableData"
                :cell-style="cellStyle"
                style="width: 100%">
                <el-table-column
                  prop="commitId"
                  label="CommitId"
                  width="330">
                </el-table-column>
                <el-table-column
                  prop="author"
                  label="author"
                  width="120">
                </el-table-column>
                <el-table-column
                  prop="comment"
                  label="comment"
                  width="">
                </el-table-column>
              </el-table>
              <el-button style="margin-top: 8px; border-radius: 0px; padding: 5px 15px;" type="primary" size="mini">重新构建</el-button>
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
              <el-button style="margin-top: 8px; border-radius: 0px; padding: 5px 15px;" type="primary" size="mini">重新构建</el-button>
              <el-button style="margin-top: 8px; border-radius: 0px; padding: 5px 15px;" type="primary" size="mini">重新执行</el-button>
            </template>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { getPipeline } from '@/api/pipeline/pipeline'
import { listBuilds, buildPipeline } from '@/api/pipeline/build'

export default {
  name: 'PipelineWorkspace',
  components: {
    Clusterbar
  },
  data() {
    return {
      titleName: [],
      search_name: '',
      users: [],
      cellStyle: {border: 0, padding: '1px 0', 'line-height': '35px'},
      maxHeight: window.innerHeight - 145,
      loading: true,
      pipeline: [],
      builds: [],
      tableData: [{
            commitId: '64a986150874cd1ed1c984889229b1204c9503d1',
            author: 'lizeen',
            comment: 'add helm application and crd'
          },],
      buildDetails: {}
    }
  },
  created() {
    this.fetchPipeline();
    this.fetchBuilds();
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
      getPipeline(this.pipelineId)
        .then((response) => {
          this.pipeline = response.data || {};
          if (this.pipeline){
            this.titleName = ["流水线", this.pipeline.pipeline.name]
          }
        }).catch(() => {
          
        })
    },
    fetchBuilds() {
      this.loading = true
      listBuilds(this.pipelineId)
        .then((response) => {
          this.loading = false
          this.builds = response.data || [];
        })
        .catch(() => {
          this.loading = false
        })
    },
    nameClick: function(id) {
      this.$router.push({name: "pipelineBuilds", params: { pipelineId: id },});
    },
    nameSearch: function(val) {
      this.search_name = val
    },
    buildPipeline: function() {
      buildPipeline(this.pipelineId, {'branch': 'develop'}).then((response) => {
        this.$message({message: '构建成功', type: 'success'});
        this.fetchBuilds()
      }).catch( (err) => {
        console.log(err)
      })
    },
    getStageStatus(status) {
      if(status == 'ok') return 'success'
      if(status == 'error') return 'error'
      if(status == 'wait') return 'wait'
      if(status == 'doing') return 'process'
      if(status == 'cancel') return 'wait'
    },
    getBuildStatusColor(status) {
      if(status == 'ok') return '#67c23a'
      if(status == 'error') return '#DC143C'
      if(status == 'doing') return '#E6A23C'
      return ''
    },
    getStageExecTime(execTimeStr, endTimeStr) {
      console.log(execTimeStr, endTimeStr)  
      if(!endTimeStr) var endTime = new Date();
      else var endTime = new Date(endTimeStr)

      var execTime = new Date(execTimeStr)
      var diffTime = endTime.getTime()-execTime.getTime()
      var stageTime = ''

      var days = Math.floor(diffTime / (24*3600*1000))
      if (days) stageTime = days + 'd'

      var leave1 = diffTime % (24*3600*1000)
      var hours = Math.floor(leave1/(3600*1000))
      if(hours) stageTime += hours + 'h'

      var leave2=leave1%(3600*1000)        //计算小时数后剩余的毫秒数
      var minutes=Math.floor(leave2/(60*1000))
      if(minutes) stageTime += minutes + 'm'

      var leave3 = leave2 % (60*1000)
      var seconds=Math.round(leave3/1000)
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
      this.$set(build, 'clickDetail', {type: type, stage: stage})
    },
    clickBuildNumber(build) {
      this.$router.push({name: 'pipelineBuildDetail', params: {buildId: build.pipeline_run.id}})
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
  margin: 15px 20px;
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
</style>
<style>
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
</style>
