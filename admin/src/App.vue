<script setup>
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { DataAnalysis, Expand, Fold, List, Operation, Setting, Tools, User, UserFilled } from "@element-plus/icons-vue";
import { clearSession, getUser } from "./auth/session";

const route = useRoute();
const router = useRouter();
const collapsed = ref(false);
const activePath = computed(() => route.path);
const currentUser = computed(() => getUser());

function navigate(path) {
  router.push(path);
}

function logout() {
  clearSession();
  router.replace("/login");
}

window.addEventListener("admin-session-expired", () => router.replace("/login"));
</script>

<template>
  <router-view v-if="route.meta.public" />
  <el-container v-else class="app-shell">
    <el-aside :width="collapsed ? '68px' : '224px'" class="app-sidebar">
      <div class="brand" :class="{ compact: collapsed }">
        <span class="brand-mark">FP</span>
        <div v-if="!collapsed" class="brand-copy">
          <strong>翻牌消除</strong>
          <span>运营管理后台</span>
        </div>
      </div>

      <el-menu
        :default-active="activePath"
        :collapse="collapsed"
        :collapse-transition="false"
        class="sidebar-menu"
        @select="navigate"
      >
        <el-menu-item index="/dashboard">
          <el-icon><DataAnalysis /></el-icon><template #title>运营概览</template>
        </el-menu-item>
        <el-menu-item index="/levels">
          <el-icon><List /></el-icon><template #title>关卡管理</template>
        </el-menu-item>
        <el-menu-item index="/players">
          <el-icon><User /></el-icon><template #title>用户管理</template>
        </el-menu-item>
        <el-menu-item index="/ads"><el-icon><Setting /></el-icon><template #title>广告频控</template></el-menu-item>
        <el-sub-menu index="system">
          <template #title><el-icon><Tools /></el-icon><span>系统管理</span></template>
          <el-menu-item index="/system/admins"><el-icon><UserFilled /></el-icon><template #title>管理员管理</template></el-menu-item>
          <el-menu-item index="/system/controls"><el-icon><Operation /></el-icon><template #title>特殊配置</template></el-menu-item>
        </el-sub-menu>
      </el-menu>
    </el-aside>

    <el-container class="workspace">
      <el-header class="topbar">
        <el-button text circle :aria-label="collapsed ? '展开导航' : '收起导航'" @click="collapsed = !collapsed">
          <el-icon size="19"><Expand v-if="collapsed" /><Fold v-else /></el-icon>
        </el-button>
        <el-breadcrumb separator="/">
          <el-breadcrumb-item>翻牌消除</el-breadcrumb-item>
          <el-breadcrumb-item>{{ route.meta.title }}</el-breadcrumb-item>
        </el-breadcrumb>
        <div class="topbar-spacer" />
        <el-tag type="success" effect="plain" round>服务已连接</el-tag>
        <el-dropdown trigger="click">
          <div class="admin-profile"><el-avatar :size="32">{{ currentUser?.displayName?.slice(0, 1) || '管' }}</el-avatar><span>{{ currentUser?.displayName || currentUser?.username }}</span></div>
          <template #dropdown><el-dropdown-menu><el-dropdown-item disabled>{{ currentUser?.role }}</el-dropdown-item><el-dropdown-item divided @click="logout">退出登录</el-dropdown-item></el-dropdown-menu></template>
        </el-dropdown>
      </el-header>
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>
