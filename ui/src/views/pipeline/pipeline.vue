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
        <el-table-column
          prop="create_time"
          label="最新构建"
          show-overflow-tooltip
        >
          <template slot-scope="scope">
            <span>
              {{ scope.row.last_build ? '#' + scope.row.last_build.build_number : "无" }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="create_time"
          label="最新构建时间"
          show-overflow-tooltip
        >
          <template slot-scope="scope">
            <span>
              {{ scope.row.last_build ? $dateFormat(scope.row.last_build.create_time) : "" }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="create_time"
          label="触发人"
          show-overflow-tooltip
        >
          <template slot-scope="scope">
            <span>
              {{ scope.row.last_build ? scope.row.last_build.operator : "" }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="status"
          label="状态"
          min-width="34%"
          show-overflow-tooltip
        >
          <template slot-scope="scope">
            <template v-if="scope.row.last_build">
              <span :style="{'color': (scope.row.last_build.status === 'ok' ? '#409EFF' : '#F56C6C')}">
                {{scope.row.status}}
                <template v-if="scope.row.last_build.status === 'ok'">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="correct" />
                </template>
                <template v-else>
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="wrong" />
                </template>
              </span>
            </template>
          </template>
        </el-table-column>

        <el-table-column label="" width="80">
          <template slot-scope="scope">
            <el-dropdown size="medium" >
              <el-link :underline="false"><svg-icon style="width: 1.3em; height: 1.3em;" icon-class="operate" /></el-link>
              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item @click.native.prevent="console.log(scope.row)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="delete" />
                  <span style="margin-left: 5px;">构建</span>
                </el-dropdown-item>
                <el-dropdown-item @click.native.prevent="">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="delete" />
                  <span style="margin-left: 5px;">编辑</span>
                </el-dropdown-item>
                <el-dropdown-item @click.native.prevent="">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="delete" />
                  <span style="margin-left: 5px;">删除</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
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
    createPipelineFunc() {

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
