<template>
  <div class="side-bar">
    <div v-if="hasTopSelector" class="top-selector">
      <div class="selector-header">
        <div class="selector-header__left">
          <i class="el-icon-arrow-left"></i>
        </div>
        <div>
          <el-select v-model="topSelectorValue" placeholder=""  @change="topSelectorChange">
            <el-option
              v-for="item in oriTopSelectors"
              :key="item.value"
              :label="item.label"
              :value="item.value"
              :disabled="item.disabled">
            </el-option>
          </el-select>
        </div>
      </div>
    </div>
    <div :class="hasTopSelector? 'scrollbar-div-wrapper-hastop' : 'scrollbar-div-wrapper-notop'">
    <el-scrollbar wrap-class="scrollbar-wrapper">
      <el-menu
        :default-active="activeMenu"
        :collapse="false"
        :unique-opened="false"
        :collapse-transition="true"
        mode="vertical"
      >
        <sidebar-item v-for="route in routes" :key="route.path" :item="route" />
      </el-menu>
    </el-scrollbar>
    </div>
  </div>
</template>

<script>
import SidebarItem from './SidebarItem'
import { listWorkspaces } from '@/api/pipeline/workspace'
import { listCluster } from '@/api/cluster'

export default {
  data() {
    return {
      oriTopSelectors: [],
      topSelectorValue: "",
    }
  },
  created() {
    this.fetchTopSelectors()
  },
  components: { SidebarItem },
  watch: {
    topSelectorType: function () {
      this.fetchTopSelectors()
    }
  },
  computed: {
    routes() {
      return this.$router.options.routes
    },
    activeMenu() {
      const route = this.$route
      const { meta, name } = route
      if (meta.activeMenu) {
        return meta.activeMenu
      }
      if (meta && meta.sideName) return meta.sideName
      return name
    },
    hasTopSelector() {
      if (["workspace", "pipeline", "cluster"].indexOf(this.topSelectorType) > -1) return true
      return false
    },
    topSelectorType() {
      const route = this.$route
      const { meta } = route
      return meta ? meta.group : ''
    },
  },
  methods: {
    fetchTopSelectors() {
      if (this.hasTopSelector) {
        if(this.topSelectorType == 'pipeline') {
          this.fetchPipelineWorkspaces()
        } else if (this.topSelectorType == 'cluster') {
          this.fetchClusters()
        } else if (this.topSelectorType == 'workspace') {
          this.fetchPipelineWorkspaces()
        }
      }
    },
    fetchClusters() {
      listCluster().then(response => {
        var clusters = response.data ? response.data : []
        var cur_cluster = this.$route.params.name
        for (let cluster of clusters) {
          let x = {'value': cluster.name, 'label': cluster.name}
          if(cluster.status != 'Connect') x['disabled'] = true;
          this.oriTopSelectors.push(x)
          if (cluster.name == cur_cluster) {
            this.topSelectorValue = cur_cluster
          }
        }
      }).catch(() => {
        
      })
    },
    fetchPipelineWorkspaces() {
      listWorkspaces().then((response) => {
        var workspaces = response.data ? response.data : []
        var cur_workspace = parseInt(this.$route.params.workspaceId)
        for (let workspace of workspaces) {
          this.oriTopSelectors.push({'value': workspace.id, 'label': workspace.name})
          if(workspace.id == cur_workspace) {
            this.topSelectorValue = cur_workspace
          }
        }
      }).catch(() => {

      })
    },
    topSelectorChange() {
      if (this.topSelectorType == 'cluster') {
        this.$router.push({name: 'cluster', params: {'name': this.topSelectorValue}})
      } else if (this.topSelectorType == 'pipeline') {
        this.$router.push({name: 'pipeline', params: {'workspaceId': this.topSelectorValue}})
      } else if (this.topSelectorType == 'workspace') {
        this.$router.push({name: 'workspaceOverview', params: {'workspaceId': this.topSelectorValue}})
      }
    }
  }
}
</script>

<style lang="scss" scoped>
  .top-selector {
    line-height: 45px;
    height: 45px; 
    position:relative; 
    border-bottom: 1px solid #EBEEF5;
    box-sizing: border-box;
    padding-left: 24px;

    .el-page-header {
      display: inline-flex;
      vertical-align: middle;

    }

    .selector-header {
      line-height: 24px;
      box-sizing: border-box;
      vertical-align: middle;
      display: inline-flex;

      .selector-header__left {
        display: flex;
        box-sizing: border-box;
        cursor: pointer;

        .el-icon-arrow-left {
          font-size: 18px;
          margin-right: 6px;
          align-self: center;
        }
      }
    }
  }

  .scrollbar-div-wrapper-hastop {
    height: calc(100% - 46px);
  }
  .scrollbar-div-wrapper-notop {
    height: 100%;
  }
</style>
<style>
  .top-selector .el-page-header__left {
    margin-right: 30px;
  }
  .top-selector .el-page-header__left::after {
    right: -15px;
  }
  .top-selector .el-input__inner {
    border: 0px;
    padding-left: 1px;
  }
</style>
