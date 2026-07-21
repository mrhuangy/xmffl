import { createRouter, createWebHistory } from "vue-router";
import { isAuthenticated } from "../auth/session";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", redirect: () => isAuthenticated() ? "/dashboard" : "/login" },
    { path: "/login", name: "login", component: () => import("../views/LoginView.vue"), meta: { public: true, title: "管理员登录" } },
    { path: "/dashboard", name: "dashboard", component: () => import("../views/DashboardView.vue"), meta: { title: "运营概览" } },
    { path: "/players", name: "players", component: () => import("../views/PlayersView.vue"), meta: { title: "用户管理" } },
    { path: "/levels", name: "levels", component: () => import("../views/LevelsView.vue"), meta: { title: "关卡管理" } },
    { path: "/ads", name: "ads", component: () => import("../views/AdConfigView.vue"), meta: { title: "广告频控" } },
    { path: "/system/admins", name: "admins", component: () => import("../views/AdminsView.vue"), meta: { title: "管理员管理" } },
    { path: "/system/controls", name: "system-controls", component: () => import("../views/SystemControlsView.vue"), meta: { title: "特殊配置" } }
  ]
});

router.beforeEach((to) => {
  const loggedIn = isAuthenticated();
  if (!to.meta.public && !loggedIn) return { name: "login", query: { redirect: to.fullPath } };
  if (to.name === "login" && loggedIn) return { name: "dashboard" };
});

export default router;
