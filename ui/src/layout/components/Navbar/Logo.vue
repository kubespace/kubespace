<template>
  <div class="navbar-logo-container">
    <transition name="navbarLogoFade">
      <router-link key="logo" class="navbar-logo-link" to="/">
        <!-- <img v-if="logo" :src="logo" class="navbar-logo"> -->
        <h1 class="navbar-title"><span style="color:#409eff">Kube</span>Space</h1>
      </router-link>
    </transition>
  </div>
</template>

<script>
export default {
  name: 'NavbarLogo',
  data() {
    return {
      title: 'KubeSpace',
    }
  }
}
</script>

<style lang="scss" scoped>
.navbarLogoFade-enter-active {
  transition: opacity 1.5s;
}

.navbarLogoFade-enter,
.navbarLogoFade-leave-to {
  opacity: 0;
}

.navbar-logo-container {
  display: inline-block;
  height: 50px;
  line-height: 50px;
  overflow: hidden;
  min-width: 130px;
  vertical-align: top;

  & .navbar-logo-link {
    height: 100%;

    & .navbar-title {
      margin: 0px 0px 0px 25px;
      font-weight: 500;
      font-size: 20px;
      font-family: Helvetica Neue, Arial, Helvetica, sans-serif;
      vertical-align: top;
    }
  }

}
</style>
