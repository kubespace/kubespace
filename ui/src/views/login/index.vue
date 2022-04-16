<template>
    <div class="login">
        <div class="login-box">
            <el-row style="margin-bottom: 25px;">
                <el-col :span="16" :offset="4" align="center">
                    <span class='login-title'>KubeSpace</span>
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
    top: 0;
    right: 0;
    bottom: 0;
    left: 0;
    width: 450px;
}
.login-box {
    border-radius: 15px;
    padding: 40px 0px 25px;
    margin: 100px 0px;
    box-shadow: 0 2px 12px 0 rgba(191, 194, 201, 0.733);
    /* box-shadow: 0 2px 4px rgba(0, 0, 0, .12), 0 0 6px rgba(0, 0, 0, .04) */
}
.el-row {
    margin-bottom: 20px;
}
.el-input__icon {
    font-size: 20px;
}
.login-title {
    font-size: 32px;
    font-family: Avenir, Helvetica Neue, Arial, Helvetica, sans-serif;
    /* background: linear-gradient(to right,#F56C6C, rgb(110, 147, 184)); */
    /* -webkit-background-clip: text; */
    /* color: transparent; */
}
</style>