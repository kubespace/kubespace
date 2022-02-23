<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :createFunc="createPipelineFunc" createDisplay="创建流水线"/>
    <div class="dashboard-container" ref="tableCot">
      <el-table
        :data="pipelines"
        class="table-fix"
        tooltip-effect="dark"
        :max-height="maxHeight"
        style="width: 100%"
        v-loading="loading"
        :cell-style="cellStyle"
        :default-sort = "{prop: 'name'}"
        row-key="name"
      >
        <el-table-column prop="name" label="名称" show-overflow-tooltip>
          <template slot-scope="scope">
            <span class="name-class" v-on:click="nameClick(scope.row.pipeline.id)">
              {{ scope.row.pipeline.name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="最新构建" show-overflow-tooltip>
          <template slot-scope="scope">
            <span v-if="scope.row.last_build" :style="{'color': getBuildStatusColor(scope.row.last_build.status)}">
              <i style="font-size: 16px;" :class="getBuildStatusIcon(scope.row.last_build.status)"></i>
              {{ scope.row.last_build ? '#' + scope.row.last_build.build_number : "无" }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="构建分支" show-overflow-tooltip>
          <template slot-scope="scope">
            <span>
              {{ getBuildBranch(scope.row.last_build) || '—' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="最新构建时间" show-overflow-tooltip>
          <template slot-scope="scope">
            <span>
              {{ scope.row.last_build ? $dateFormat(scope.row.last_build.create_time) : "" }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="触发人" show-overflow-tooltip>
          <template slot-scope="scope">
            <span>
              {{ scope.row.last_build ? scope.row.last_build.operator : "" }}
            </span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="140">
          <template slot-scope="scope">
            <div class="tableOperate">
              <el-link :underline="false" style="margin-right: 15px; color:#409EFF" @click="nameClick(scope.row.id)">构建</el-link>
              <el-link :underline="false" style="margin-right: 15px; color:#409EFF" @click="editPipelineOperate(scope.row.pipeline.id)">编辑</el-link>
              <el-link :underline="false" style="color: #F56C6C" @click="handleDeleteWorkspace(scope.row.id, scope.row.name)">删除</el-link>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { listPipelines } from '@/api/pipeline/pipeline'
import { Message } from 'element-ui'

export default {
  name: 'PipelineWorkspace',
  components: {
    Clusterbar
  },
  data() {
    return {
      titleName: ["流水线"],
      search_name: '',
      users: [],
      cellStyle: {border: 0},
      maxHeight: window.innerHeight - 150,
      loading: true,
      pipelines: [],
      createClusterFormVisible: false,
      inviteForm: false,
      clusterConnectDialog: false,
      clusterConnectToken: '',
      form: {
        name: '',
        members: [],
      },
      locationAddr: window.location.origin,
    }
  },
  created() {
    this.fetchData();
  },
  mounted() {
    const that = this
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - 150
        console.log(heightStyle)
        that.maxHeight = heightStyle
      })()
    }
  },
  computed: {
    workspaceId() {
      return this.$route.params.workspaceId
    }
  },
  methods: {
    fetchData() {
      this.loading = true
      listPipelines(this.workspaceId)
        .then((response) => {
          this.loading = false
          this.pipelines = response.data || [];
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
    getBuildBranch(build) {
      console.log(build)
      if(build) {
        let branch = ''
        let author = ''
        if(build.env && build.env.PIPELINE_CODE_BRANCH) {
          branch = build.env.PIPELINE_CODE_BRANCH
        }
        if (!branch) return ''
        if(build.env && build.env.PIPELINE_CODE_COMMIT_AUTHOR){
          author = build.env.PIPELINE_CODE_COMMIT_AUTHOR
        }
        return branch + (author ? ' (' + author + ')' : '')
      }
      return ''
    },
    createPipelineFunc() {
      this.$router.push({name: "pipelineCreate",});
    },
    editPipelineOperate(id) {
      this.$router.push({name: "pipelineEdit", params: { pipelineId: id },});
    },
    getBuildStatusColor(status) {
      if(status == 'ok') return '#67c23a'
      if(status == 'error') return '#DC143C'
      if(status == 'doing') return '#E6A23C'
      return ''
    },
    getBuildStatusIcon(status) {
      if(status == 'ok') return 'el-icon-success'
      if(status == 'error') return 'el-icon-error'
      if(status == 'doing') return 'el-icon-refresh'
      if(status == 'wait') return 'el-icon-remove'
      return ''
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
</style>
