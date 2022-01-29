<template>
  <div>
    <template v-if="item.hidden">
      <template v-if="item.children && item.children.length > 0">
        <sidebar-item
          v-for="child in item.children"
          :key="child.name"
          :item="child"
        />
      </template>
    </template>
    <template v-else-if="item.meta && item.meta.group && item.meta.group == pathGroup">
      <template v-if="item.children && item.children.length > 0">
        <template v-if="hasPerm(item)">
          <el-submenu ref="subMenu" :index="item.name" popper-append-to-body>
            <template slot="title">
              <item v-if="item.meta" :icon="item.meta && item.meta.icon" :title="item.meta.title" />
            </template>
            <sidebar-item
              v-for="child in item.children"
              :key="child.name"
              :item="child"
            />
          </el-submenu>
        </template>
      </template>
      <template v-else-if="hasPerm(item)">
        <span v-on:click="routeTo(item)">
          <el-menu-item :index="item.name">
            <item :icon="item.meta && item.meta.icon" :title="item.meta.title" />
          </el-menu-item>
        </span>
      </template>
    </template>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import Item from './Item'
import { hasPermission } from "@/api/settings_role";

export default {
  name: 'SidebarItem',
  components: { Item },
  props: {
    // route object
    item: {
      type: Object,
      required: true
    },
  },
  data() {
    return {}
  },
  computed: {
    pathGroup() {
      const meta = this.$route.meta
      var group = ''
      if (meta && meta.group) {
        group = meta.group
      }
      return group
    }
  },
  methods: {
    routeTo(item) {
      const route = this.$route
      this.$router.push({name: item.name, params: route.params})
    },
    hasPerm(item) {
      if(item.children && item.children.length > 0) {
        for(var childItem of item.children) {
          var meta = childItem.meta
          if(meta.perm) return true
          if(hasPermission(meta.group, meta.object, 'get')) {
            return true
          }
        }
        return false
      } else {
        var meta = item.meta
        if(meta.perm) return true
        return hasPermission(meta.group, meta.object, 'get')
      }
    }
  }
}
</script>

<style >
.sidebar-container .el-menu-item:focus {
  background-color: rgba(255,255,255,0);
}
.sidebar-container .el-menu-item:hover {
  background-color: #ecf5ff;
}
.sidebar-container .el-menu-item {
  height: 45px;
  line-height: 45px;
}
.sidebar-container .el-submenu__title {
  height: 45px;
  line-height: 45px;
}
.sidebar-container .el-submenu__icon-arrow {
  margin-top: -5px;
}
</style>
