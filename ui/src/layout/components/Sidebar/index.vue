<template>
  <div class="side-bar">
    <div v-if="hasTopSelector" class="top-selector">
      <div class="selector-header">
        <div class="selector-header__left" @click="topSelectorReturn">
          <i class="el-icon-arrow-left"></i>
        </div>
        <div class="topSelector__right">
          <el-select v-model="topSelectorValue" v-loading="topSelectorLoading"  placeholder=""  @change="topSelectorChange">
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
import { listProjects } from "@/api/project/project";

export default {
  data() {
    return {
      oriTopSelectors: [],
      topSelectorValue: "",
      topSelectorLoading: true
    }
  },
  created() {
    if (this.hasTopSelector && this.topSelectorType == 'cluster') {
      this.$store.dispatch('watchCluster', this.$route.params.clusterId)
      this.$store.dispatch('watchNamespace', '')
    }
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
          this.fetchProjects()
        }
      } else {
        this.$store.dispatch('watchCluster', '')
        this.$store.dispatch('watchNamespace', '')
      }
    },
    fetchClusters() {
      this.topSelectorLoading = true;
      listCluster().then(response => {
        var clusters = response.data ? response.data : []
        clusters.sort(function(a,b) {
          return a.status.localeCompare(b.status)
        })
        var cur_cluster = this.$route.params.clusterId
        for (let cluster of clusters) {
          let x = {'value': cluster.name, 'label': cluster.name1}
          if(cluster.status != 'Connect') x['disabled'] = true;
          this.oriTopSelectors.push(x)
          if (cluster.name == cur_cluster) {
            this.topSelectorValue = cur_cluster
          }
        }
        this.topSelectorLoading = false;
      }).catch(() => {
        
      })
    },
    fetchPipelineWorkspaces() {
      this.topSelectorLoading = true;
      listWorkspaces().then((response) => {
        var workspaces = response.data ? response.data : []
        var cur_workspace = parseInt(this.$route.params.workspaceId)
        for (let workspace of workspaces) {
          this.oriTopSelectors.push({'value': workspace.id, 'label': workspace.name})
          if(workspace.id == cur_workspace) {
            this.topSelectorValue = cur_workspace
          }
        }
        this.topSelectorLoading = false;
      }).catch(() => {

      })
    },
    fetchProjects() {
      this.topSelectorLoading = true;
      listProjects().then((response) => {
        var projects = response.data ? response.data : []
        var cur_workspace = parseInt(this.$route.params.workspaceId)
        for (let workspace of projects) {
          this.oriTopSelectors.push({
            'value': workspace.id, 'label': workspace.name, cluster_id: workspace.cluster_id, namespace: workspace.namespace,
          })
          if(workspace.id == cur_workspace) {
            this.topSelectorValue = cur_workspace
            this.$store.dispatch('watchCluster', workspace.cluster_id)
            this.$store.dispatch('watchNamespace', workspace.namespace)
          }
        }
        this.topSelectorLoading = false;
      }).catch(() => {

      })
    },
    topSelectorChange() {
      if (this.topSelectorType == 'cluster') {
        // this.$router.push({name: 'cluster', params: {'name': this.topSelectorValue}})
        let changeRoute = this.$route.name
        console.log("top selector change ", changeRoute)
        if(this.$route.meta.sideName) {
          changeRoute = this.$route.meta.sideName
        }
        this.$store.dispatch('watchCluster', this.topSelectorValue)
        this.$store.dispatch('watchNamespace', '')
        this.$router.push({name: changeRoute, params: {'clusterId': this.topSelectorValue}})
      } else if (this.topSelectorType == 'pipeline') {
        this.$router.push({name: 'pipeline', params: {'workspaceId': this.topSelectorValue}})
      } else if (this.topSelectorType == 'workspace') {
        for(let s of this.oriTopSelectors) {
          if(s.value == this.topSelectorValue) {
            this.$store.dispatch('watchCluster', s.cluster_id)
            this.$store.dispatch('watchNamespace', s.namespace)
          }
        }
        let changeRoute = this.$route.name
        if(this.$route.meta.sideName) {
          changeRoute = this.$route.meta.sideName
        }
        // this.$router.push({name: 'workspaceOverview', params: {'workspaceId': this.topSelectorValue}})
        this.$router.push({name: changeRoute, params: {'workspaceId': this.topSelectorValue}})
        
      }
    },
    topSelectorReturn() {
      if (this.topSelectorType == 'cluster') {
        this.$router.push({name: 'clusterIndex'})
      } else if (this.topSelectorType == 'pipeline') {
        this.$router.push({name: 'pipelineWorkspace'})
      } else if (this.topSelectorType == 'workspace') {
        this.$router.push({name: 'workspaceIndex'})
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
  .top-selector .el-input__inner {
    border: 0px;
    padding-left: 1px;
  }

  .topSelector__right .el-loading-spinner {
    margin-top: -10px;
  }
  .topSelector__right .el-loading-spinner .circular {
    height: 22px;
    width: 22px;
    margin-right: 65px;
  }
</style>
