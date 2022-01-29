<template>
    <div class="login">
        <div class="login-box">
            <el-row style="margin-bottom: 25px;">
                <el-col :span="16" :offset="4" align="center">
                    <span class='login-title'>OpenSpace</span>
                </el-col>
            </el-row>
            <el-row style="margin-bottom: 8px">
                <el-col :span="16" :offset="4">
                    <el-input 
                        placeholder="请输入用户名" name="username" tabindex="1" disabled
                        v-model="username" class="login-input" auto-complete="on">
                        <i slot="prefix" class="el-input__icon el-icon-user"></i>
                    </el-input>
                </el-col>
            </el-row>
            <el-row style="margin-bottom: 8px">
                <el-col :span="16" :offset="4">
                <span class="span-text">第一次登录，请输入admin管理员密码</span>
                </el-col>
            </el-row>
            <el-row>
                <el-col :span="16" :offset="4">
                    <el-input type="password" autocomplete="new-password" name="password" tabindex="2"
                        placeholder="请输入密码" clearable
                        v-model="password" >
                        <i slot="prefix" class="el-input__icon el-icon-lock"></i>
                    </el-input>
                </el-col>
            </el-row>
            <el-row>
                <el-col :span="16" :offset="4">
                    <el-input type="password" autocomplete="confirm-password" name="confirmPassword" tabindex="3"
                        placeholder="请再次输入密码" clearable
                        v-model="confirmPassword" @keyup.enter.native="handleAdmin">
                        <i slot="prefix" class="el-input__icon el-icon-lock"></i>
                    </el-input>
                </el-col>
            </el-row>
            <el-row style="margin-top: 30px;">
                <el-col :span="16" :offset="4" align="center">
                    <el-button :loading="loading" size="medium" plain style="width: 100%;" @click.native.prevent="handleAdmin">
                        <span style="margin-right: 20px;">确</span>
                        <span>认</span>
                    </el-button>
                </el-col>
            </el-row>
        </div>
    </div>
</template>

<script>
import { Message } from 'element-ui'
import { adminSet } from '@/api/user'

export default {
    data() {
        return {
            username: 'admin',
            password: '',
            confirmPassword: '',
            loading: false,
        }
    },
    methods: {
        handleAdmin() {
            const password = this.password
            const confirmPassword = this.confirmPassword
            if (!password) {
                Message.error("密码不能为空！")
                return false
            }
            if (!confirmPassword) {
                Message.error("确认密码不能为空！")
                return false
            }
            if (password !== confirmPassword) {
                Message.error("两次输入密码不同，请重新输入！")
                return false
            }
            const adminPassword = {password: password}
            this.loading = true
            adminSet(adminPassword).then(() => {
                this.loading = false
                parent.location.href = "/ui/login"
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
}
.span-text {
    font-size: 13px;
    color: rgba(86, 88, 92, 0.733);
    font-family: Avenir, Helvetica Neue, Arial, Helvetica, sans-serif;
}
</style>