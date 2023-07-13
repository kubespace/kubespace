<template>
<div style="background-color: #fff; width:100%; height: 100%;">
    <div style="background-color: #fbfbfb; width:55%; height: 100%;">
      <section class="section-class">
        <div class="center">
          <div style="font-size: 48px; font-weight: 400; margin: 30px 0px 20px;">
            KubeSp<svg-icon class="icon-class" icon-class="kubespace"/>ce
          </div>
          <p class="center-content">致力于提升DevOps效能的Kubernetes多集群管理平台</p>
        </div>
      </section>
    </div>
    <div class="login">
        <div class="login-box">
            <el-row style="margin-bottom: 25px;">
                <el-col :span="16" :offset="4" align="center">
                    <span class='login-title'><span style="color: #409eff">Kube</span>Space</span>
                </el-col>
            </el-row>
            <el-row>
                <el-col :span="16" :offset="4">
                    <el-input 
                        placeholder="请输入用户名" name="username" tabindex="1"
                        v-model="username" class="login-input" clearable autofocus>
                        <i slot="prefix" class="el-input__icon el-icon-user"></i>
                    </el-input>
                </el-col>
            </el-row>
            <el-row>
                <el-col :span="16" :offset="4">
                    <el-input type="password" autocomplete="new-password" name="password" tabindex="2"
                        placeholder="请输入密码" clearable
                        v-model="password" @keyup.enter.native="handleLogin">
                        <i slot="prefix" class="el-input__icon el-icon-lock"></i>
                    </el-input>
                </el-col>
            </el-row>
            <el-row style="margin-top: 30px;">
                <el-col :span="16" :offset="4" align="center">
                    <el-button :loading="loading" size="medium" plain style="width: 100%;" @click.native.prevent="handleLogin">
                        <span style="margin-right: 20px;">登</span>
                        <span>录</span>
                    </el-button>
                </el-col>
            </el-row>
        </div>
    </div>
</div>
</template>

<script>
import { Message } from 'element-ui'
import { hasAdmin } from '@/api/user'

export default {
    data() {
        return {
            username: '',
            password: '',
            loading: false,
            redirect: undefined
        }
    },
    beforeRouteEnter(to, from, next) {
        hasAdmin().then(response => {
            const { data } = response
            const { has } = data
            if (has) {
                next()
            } else {
                next('/ui/login/admin')
            }
        }).catch(() => {
            next()
        })
    },
    watch: {
        $route: {
            handler: function(route) {
                this.redirect = route.query && route.query.redirect
            },
            immediate: true
        }
    },
    methods: {
        handleLogin() {
            const userInfo = {username: this.username, password: this.password}
            console.log(userInfo)
            if (!userInfo.username) {
                Message.error("用户名不能为空！")
                return false
            }
            if (!userInfo.password) {
                Message.error("密码不能为空！")
                return false
            }
            this.loading = true
            this.$store.dispatch('user/login', userInfo).then(() => {
                this.loading = false
                // console.log(this.redirect)
                parent.location.href = this.redirect || '/'
                // this.$router.push({ path: this.redirect || '/' })
            }).catch(() => {
                this.loading = false
            })
        }
    }
}
</script>

<style scoped>
.login {
    margin: auto;
    position: absolute;
    top: 10%;
    right: 0;
    bottom: 0;
    left: 55%;
    width: 450px;
}
.login-box {
    border-radius: 15px;
    padding: 40px 0px 25px;
    margin: 100px 0px;
    box-shadow: 0px 2px 10px 0px rgba(191, 194, 201, 0.733);
    /* box-shadow: 0 2px 4px rgba(0, 0, 0, .12), 0 0 6px rgba(0, 0, 0, .04) */
}
.el-row {
    margin-bottom: 20px;
}
.el-input__icon {
    font-size: 20px;
}
.login-title {
    letter-spacing:0px;
    font-size: 29px;
    font-family: Helvetica Neue, Arial, Helvetica, sans-serif;
    /* background: linear-gradient(to right,#F56C6C, rgb(110, 147, 184)); */
    /* -webkit-background-clip: text; */
    /* color: transparent; */
}
.section-class {
    width: 100%;
    margin: 0 auto;
    padding: 0;
    box-sizing: border-box;
    text-align: center;
    z-index: 20;
    border-bottom: solid 1px #ddd;
    height: 100vh;
    position: relative;
}

.center {
    position: absolute;
    width: 100%;
    /* padding-top: 70px;   */
    top: 200px;
    min-height: 200px;
}

.center-content {
    font-family: raleway,century gothic,texgyreadventor,sans-serif;
    font-size: 12px;
    font-weight: 500;
    line-height: 22px;
    letter-spacing: 6px;
    text-transform: uppercase;
    color: #999;
    text-align: center;
}
.icon-class {
    font-size: 34px;
    animation: loading-rotate 5s linear infinite;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
}
</style>