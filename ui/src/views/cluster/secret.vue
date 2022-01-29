<template>
  <div>
    <clusterbar
      :titleName="titleName"
      :nsFunc="nsSearch"
      :nameFunc="nameSearch"
    />
    <div class="dashboard-container">
      <el-table
        ref="multipleTable"
        :data="secrets"
        class="table-fix"
        tooltip-effect="dark"
        :max-height="maxHeight"
        style="width: 100%"
        v-loading="loading"
        :cell-style="cellStyle"
        :default-sort="{ prop: 'name' }"
        row-key="uid"
      >
        <el-table-column type="selection" width="45"> </el-table-column>
        <el-table-column
          prop="name"
          label="名称"
          min-width="45"
          show-overflow-tooltip
        >
          <template slot-scope="scope">
            <span class="name-class" v-on:click="nameClick(scope.row.namespace, scope.row.name)">
              {{ scope.row.name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="namespace"
          label="命名空间"
          min-width="40"
          show-overflow-tooltip
        >
        </el-table-column>
        <el-table-column
        prop="type"
        label="Type"
        min-width="45"
        show-overflow-tooltip
      >
      </el-table-column>
        <el-table-column
          prop="create_time"
          label="创建时间"
          min-width="45"
          show-overflow-tooltip
        >
        </el-table-column>
        <el-table-column label="" show-overflow-tooltip width="45">
          <template slot-scope="scope">
            <el-dropdown size="medium">
              <el-link :underline="false"
                ><svg-icon
                  style="width: 1.3em; height: 1.3em;"
                  icon-class="operate"
              /></el-link>
              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item
                  @click.native.prevent="
                    nameClick(scope.row.namespace, scope.row.name)
                  "
                >
                  <svg-icon
                    style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em"
                    icon-class="detail"
                  />
                  <span style="margin-left: 5px;">详情</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$updatePerm()"
                  @click.native.prevent="
                    getConfigMapYaml(scope.row.namespace, scope.row.name)
                  "
                >
                  <svg-icon
                    style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em"
                    icon-class="edit"
                  />
                  <span style="margin-left: 5px;">修改</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$deletePerm()"
                  @click.native.prevent="
                    deleteConfigMaps([
                      { namespace: scope.row.namespace, name: scope.row.name },
                    ])
                  "
                >
                  <svg-icon
                    style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em"
                    icon-class="delete"
                  />
                  <span style="margin-left: 5px;">删除</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <el-dialog
      title="编辑"
      :visible.sync="yamlDialog"
      :close-on-click-modal="false"
      width="60%"
      top="55px"
    >
      <yaml v-if="yamlDialog" v-model="yamlValue" :loading="yamlLoading"></yaml>
      <span slot="footer" class="dialog-footer">
        <el-button plain @click="yamlDialog = false" size="small"
          >取 消</el-button
        >
        <el-button plain @click="updatePod()" size="small">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar, Yaml } from '@/views/components'
import { Message } from 'element-ui'
// import { listSecrets, getSecret } from '@/api/secret'
import { listSecrets } from '@/api/secret'

export default {
  name: 'Secrets',
  components: {
    Clusterbar,
    Yaml,
  },
  data() {
    return {
      titleName: ["Secrets"],
      originSecrets: [],
      search_name: '',
      search_ns: [],
      cellStyle: { border: 0 },
      maxHeight: window.innerHeight - 150,
      loading: true,
      yamlDialog: false,
      yamlNamespace: '',
      yamlName: '',
      yamlValue: '',
      yamlLoading: true,
    }
  },
  created() {
    this.fetchData()
  },
  computed: {
    secrets: function() {
      let dlist = []
      for (let p of this.originSecrets) {
        if (this.search_ns.length > 0 && this.search_ns.indexOf(p.namespace) < 0) continue
        if (this.search_name && !p.name.includes(this.search_name)) continue
        
        dlist.push(p)
      }
      return dlist
    },
  },
  methods: {
    nameClick: function(namespace, name) {
      console.log(namespace, name);
      this.$router.push({
        name: 'secretDetail',
        params: { namespace: namespace, secretName: name },
      })
    },
    nsSearch: function(vals) {
      this.search_ns = []
      for (let ns of vals) {
        this.search_ns.push(ns)
      }
    },
    nameSearch: function(val) {
      this.search_name = val
    },
    fetchData: function() {
      this.loading = true
      this.originSecrets = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listSecrets(cluster).then(response => {
          this.loading = false
          this.originSecrets = response.data
        }).catch(() => {
          this.loading = false
        })
      } else {
        this.loading = false
        Message.error("获取集群异常，请刷新重试")
      }
    },
    // getSecretYaml: function(namespace, name) {
    //   const cluster = this.$store.state.cluster
    //   console.log('xxxxxxx', namespace, name)
    //   if (!cluster) {
    //     Message.error('获取集群参数异常，请刷新重试')
    //     return
    //   }
    //   if (!namespace) {
    //     Message.error('获取命名空间参数异常，请刷新重试')
    //     return
    //   }
    //   if (!name) {
    //     Message.error('获取Secret名称参数异常，请刷新重试')
    //     return
    //   }
    //   this.yamlLoading = true
    //   this.yamlDialog = true
    //   getSecret(cluster, namespace, name, 'yaml')
    //     .then((response) => {
    //       this.yamlLoading = false
    //       this.yamlValue = response.data
    //       this.yamlNamespace = namespace
    //       this.yamlName = name
    //     })
    //     .catch(() => {
    //       this.yamlLoading = false
    //     })
    // },
  }
}
</script>

<style lang="scss" scoped>
  .dashboard {
    &-container {
      margin: 10px 30px;
    }
    &-text {
      font-size: 30px;
      line-height: 46px;
    }
  
    .table-fix {
      height: calc(100% - 100px);
    }
  }
  
  .name-class {
    cursor: pointer;
  }
  .name-class:hover {
    color: #409EFF;
  }
  </style>