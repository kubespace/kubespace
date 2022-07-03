<template>
  <div class="cluster-bar">
    <el-breadcrumb class="app-breadcrumb" separator-class="el-icon-arrow-right">
        <el-breadcrumb-item v-for="(t, i) in titleName" :key="i" class="no-redirect">
          <span v-if='linkWithTitle(i)' @click="routeLink(linkWithTitle(i))" class="linkTitleClass">{{ t }}</span>
          <span v-else>{{ t }}</span>
        </el-breadcrumb-item>
    </el-breadcrumb>
    
    <el-link v-if="typeof addFunc !== 'undefined'" class="icon-create" @click="addFunc()"><svg-icon icon-class="add"/></el-link>
    <el-link v-if="typeof saveFunc !== 'undefined'" class="icon-create" @click="saveFunc()"><svg-icon icon-class="save"/></el-link>
    <span class="icon-create">
      
    </span>
    <div class="right">
      <slot name="right-btn"></slot>
      <template>
        <el-button v-if="typeof delFunc !== 'undefined'" size="small" plain @click="delFunc()"
          type="danger" :disabled="!$editorRole()" icon="el-icon-delete" >
          删除
        </el-button>
      </template>

      <template>
        <el-button v-if="typeof editFunc !== 'undefined'" size="small" plain @click="editFunc()"
          type="success" :disabled="!$editorRole()" icon='el-icon-edit'>
          编辑
        </el-button>
      </template>

      <template>
        <el-button v-if="typeof createFunc !== 'undefined'" size="small" type="primary" @click="createFunc()"
          icon="el-icon-plus" :disabled="!$editorRole()">
          {{ createDisplay }}
        </el-button>
      </template>

      <el-select v-if="typeof nsFunc !== 'undefined'" v-model="nsInput" @change="nsChange" multiple placeholder="命名空间" size="small">
        <el-option
          v-for="item in namespaces"
          :key="item.name"
          :label="item.name"
          :value="item.name">
        </el-option>
      </el-select>
      <el-input v-if="typeof nameFunc !== 'undefined'"
        size="small"
        placeholder="搜索"
        v-model="nameInput"
        @input="nameDebounce"
        suffix-icon="el-icon-search">
      </el-input>
    </div>
  </div>
</template>

<script>
// import { mapGetters } from 'vuex'
import { listNamespace } from '@/api/namespace'
import { Message } from 'element-ui'
import storage from '@/utils/storage'

let nameTimer
export default {
  name: 'Clusterbar',
  props: {
    titleName: {
      type: Array,
      required: true,
      default: () => {return []}
    },
    titleLink: {
      type: Array,
      default: () => {return []}
    },
    nsFunc: {
      type: Function,
      required: false,
      default: undefined
    },
    nameFunc: {
      type: Function,
      required: false,
      default: undefined
    },
    delFunc: {
      type: Function,
      required: false,
      default: undefined,
    },
    addFunc: {
      type: Function,
      required: false,
      default: undefined,
    },
    saveFunc: {
      type: Function,
      required: false,
      default: undefined,
    },
    editFunc: {
      type: Function,
      required: false,
      default: undefined,
    },
    createDisplay: {
      type: String,
      required: false,
      default: "创建"
    },
    createFunc: {
      type: Function,
      required: false,
      default: undefined,
    }
  },
  data() {
    return {
      nameInput: "",
      nsInput: [],
      namespaces: [],
    }
  },
  created() {
    if (typeof this.nsFunc !== 'undefined' && this.$viewerRole()) {
      this.fetchNamespace()
    }
  },
  computed: {
    cluster: function() {
      return this.$store.state.cluster
    },
    nsKey: function() {
      return 'namespace-' + this.cluster
    }
  },
  methods: {
    routeLink(link) {
      this.$router.push({name: link})
    },
    linkWithTitle(idx) {
      return this.titleLink[idx]
    },
    nsChange(vals) {
      if (this.nsFunc) {
        this.nsFunc(vals)
      }
      if(vals && vals.length > 0) {
        storage.set(this.nsKey, vals)
      } else {
        storage.remove(this.nsKey)
      }
    },
    nameDebounce: function() {
      if (this.nameFunc) {
        if (nameTimer) {
          clearTimeout(nameTimer)
        }
        nameTimer = setTimeout(() => {
          this.nameFunc(this.nameInput)
          // this.nameModel = this.nameInput
          nameTimer = undefined
        }, 500)
      }
    },
    fetchNamespace: function() {
      this.namespaces = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listNamespace(cluster).then(response => {
          this.namespaces = response.data
          this.namespaces.sort((a, b) => {return a.name > b.name ? 1 : -1})
          let nsCache = storage.get(this.nsKey)
          if (nsCache) {
            var nsNames = []
            for(let n of this.namespaces) nsNames.push(n.name)
            let nsInput = nsCache.filter((name) => {return nsNames.indexOf(name) > -1})
            this.nsInput = nsInput
            if (this.nsFunc) {
              this.nsFunc(this.nsInput)
            }
          }
        }).catch((err) => {
          console.log(err)
        })
      } else {
        Message.error("获取集群异常，请刷新重试")
      }
    }
  }
}
</script>

<style lang="scss" scoped>
@import "~@/styles/variables.scss";
.cluster-bar {
  transition: width 0.28s;
  // height: 50px;
  overflow: hidden;
  // box-shadow: inset 0 0 4px rgba(0, 21, 41, 0.1);
  margin: 15px 30px 0px;

  .app-breadcrumb.el-breadcrumb {
    // display: inline-block;
    font-size: 14px;
    // line-height: 52px;
    //  margin-left: 8px;

    .no-redirect {
      // color: #97a8be;
      cursor: text;
      font-size: 16px;
      font-family: Avenir, Helvetica Neue, Arial, Helvetica, sans-serif;
    }
    .no-redirect:first-child {
     // margin-left: 15px;
    }
  }

  .icon-create:first {
    margin-left: 15px;
  }

  .icon-create {
    display: inline-block;
    line-height: 55px;
    margin-left: 15px;
    vertical-align: 0.95em;
    font-size: 23px;
  }

  .right {
    float: right;
    height: 100%;
    line-height: 45px;
    // margin-right: 25px;

    .el-input {
      width: 195px;
      margin-left: 15px;
    }

    .el-select {
      margin-left: 15px;
      .el-select__tags {
        white-space: nowrap;
        overflow: hidden;
      }
    }

  }
  .linkTitleClass:hover{
    cursor: pointer;
  }
}
</style>
<style >
/* .right .el-button.is-plain {
  border-color: #f78989;
  color: #f78989;
}
.right .el-button.is-plain:hover {
  border-color: #f56c6c;
  color: #f56c6c;
}
.right .el-button.is-plain:focus {
  border-color: #f56c6c;
  color: #f56c6c;
} */
</style>
