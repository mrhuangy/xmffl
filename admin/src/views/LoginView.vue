<script setup>
import { reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { Lock, User } from "@element-plus/icons-vue";
import { login } from "../api/client";
import { saveSession } from "../auth/session";

const route = useRoute();
const router = useRouter();
const formRef = ref();
const loading = ref(false);
const form = reactive({ username: "", password: "" });
const rules = {
  username: [{ required: true, message: "请输入管理员用户名", trigger: "blur" }],
  password: [{ required: true, message: "请输入密码", trigger: "blur" }]
};

async function submit() {
  await formRef.value.validate();
  loading.value = true;
  try {
    const data = await login(form);
    saveSession(data);
    ElMessage.success("登录成功");
    router.replace(typeof route.query.redirect === "string" ? route.query.redirect : "/dashboard");
  } catch (err) {
    ElMessage.error(err.message);
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <main class="login-page">
    <section class="login-panel">
      <div class="login-brand"><span class="brand-mark">FP</span><div><strong>翻牌消除</strong><span>运营管理后台</span></div></div>
      <div class="login-heading"><h1>管理员登录</h1><p>使用已授权的后台账号继续。</p></div>
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top" size="large" @keyup.enter="submit">
        <el-form-item label="用户名" prop="username"><el-input v-model.trim="form.username" :prefix-icon="User" autocomplete="username" placeholder="请输入用户名" /></el-form-item>
        <el-form-item label="密码" prop="password"><el-input v-model="form.password" :prefix-icon="Lock" type="password" show-password autocomplete="current-password" placeholder="请输入密码" /></el-form-item>
        <el-button type="primary" class="login-button" :loading="loading" @click="submit">登录</el-button>
      </el-form>
      <p class="login-footer">账号由系统管理员创建和授权</p>
    </section>
  </main>
</template>
