<template>
  <div class="navbar">
    <logo />
    <navbar-item />

    <div class="right-menu">
       <el-button v-if="importYaml" plain size="small" @click="openYamlDialog"
        style="margin-right: 25px; padding: 8px 10px;" type="info">导入YAML</el-button>
      <el-dropdown placement="bottom">
        <span class="el-dropdown-link">
          <img class="avatar-class" src="@/assets/user.png" />
          {{ username }}<i class="el-icon-arrow-down el-icon--right"></i>
        </span>
        <el-dropdown-menu slot="dropdown" @click.native="logout">
          <el-dropdown-item>登出</el-dropdown-item>
        </el-dropdown-menu>
      </el-dropdown>
    </div>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import Logo from './Logo'
import NavbarItem from './NavbarItem'
// import { Yaml } from '@/views/components'

export default {
  components: { Logo, NavbarItem,  },
  props: ['nav',],
  data() {
    return {
      importYaml: false
    }
  },

  watch: {
    cluster: function (newObj) {
      console.log(newObj)
      if(newObj) {
        this.importYaml = true
      } else {
        this.importYaml = false
      }
    }
  },

  computed: {
    ...mapGetters([
      'username',
    ]),
    cluster: function() {
      return this.$store.state.cluster
    }
  },
  methods: {
    async logout() {
      await this.$store.dispatch('user/logout')
      parent.location.href = `/ui/login`
    },
    openYamlDialog: function() {
      // this.yamlValue = "asdfwe"
      // this.yamlDialog = true;
      // this.yamlLoading = false;
      // this.value = true;
      this.nav.yamlValue = '';
      this.nav.dialog = true;
    },
  }
}
</script>

<style lang="scss" scoped>
.navbar {
  height: 50px;
  overflow: hidden;
  position: relative;
  background: #fff;
  box-shadow: 0 1px 4px rgba(0,21,41,.08);

  .right-menu {
    float: right;
    height: 100%;
    line-height: 50px;
    margin-right: 25px;

    .el-dropdown-link {
      cursor: pointer;
      // color: #409EFF;
      font-size: 16px;
      vertical-align: middle;
    }
    .el-icon-arrow-down {
      font-size: 12px;
    }

    .avatar-class {
      vertical-align: -0.25em; 
      font-size: 20px; 
      height: 20px; 
      width: 20px;
    }

  }
}
</style>
