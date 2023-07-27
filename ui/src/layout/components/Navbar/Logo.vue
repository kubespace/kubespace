<template>
  <div class="navbar-logo-container">
    <transition name="navbarLogoFade">
      <router-link key="logo" class="navbar-logo-link" to="/">
        <!-- <img v-if="logo" :src="logo" class="navbar-logo"> -->
        <svg-icon icon-class="spacelet"  class="navbarLogoIcon" style="width: 1.5em; height: 1.5em; vertical-align: -0.44em;"/>
        <h1 class="navbar-title"><span style="color:#409eff">Kube</span><span>Space</span></h1>
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

.navbarLogoIcon {
  width: 1.5em; 
  height: 1.5em; 
  display: inline-block; 
  margin-left: 25px; 
  vertical-align: -0.42em;
  // background-image: linear-gradient(to right, #eea2a2 0%, #bbc1bf 19%, #57c6e1 42%, #b49fda 79%, #7ac5d8 100%);
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
      margin: 0px 0px 0px 5px;
      font-weight: 500;
      font-size: 20px;
      font-family: Helvetica Neue, Arial, Helvetica, sans-serif;
      vertical-align: top;
      display: inline-block;
    }
  }

}
</style>
