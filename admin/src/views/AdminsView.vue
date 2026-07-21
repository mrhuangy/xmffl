<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Delete, Edit, Plus, Refresh } from "@element-plus/icons-vue";
import { createAdmin, deleteAdmin, fetchAdmins, updateAdmin } from "../api/client";
import { getUser } from "../auth/session";

const admins = ref([]);
const loading = ref(false);
const saving = ref(false);
const dialogOpen = ref(false);
const formRef = ref();
const editingId = ref(null);
const currentUser = getUser();
const currentAdmin = computed(() => admins.value.find((item) => item.username === currentUser?.username));
const form = reactive({ username: "", email: "", password: "", displayName: "", role: "operator", status: "active", permissionKeys: [] });
const roles = { owner: "超级管理员", operator: "运营管理员", viewer: "只读管理员" };
const statuses = { active: "正常", disabled: "已禁用", locked: "已锁定" };
const permissionOptions = [
  { label: "关卡管理", value: "levels" }, { label: "用户查看", value: "players" },
  { label: "广告配置", value: "ads" }, { label: "数据概览", value: "dashboard" }
];
const rules = {
  username: [{ required: true, message: "请输入用户名", trigger: "blur" }],
  displayName: [{ required: true, message: "请输入显示名称", trigger: "blur" }],
  role: [{ required: true, message: "请选择角色", trigger: "change" }],
  password: [{ validator: (_, value, callback) => {
    if (!editingId.value && !value) return callback(new Error("请输入初始密码"));
    if (value && value.length < 10) return callback(new Error("密码至少 10 位"));
    callback();
  }, trigger: "blur" }]
};

function formatDate(value) { return value ? new Intl.DateTimeFormat("zh-CN", { dateStyle: "medium", timeStyle: "short" }).format(new Date(value)) : "-"; }

async function loadAdmins() {
  loading.value = true;
  try { admins.value = (await fetchAdmins()).admins || []; }
  catch (err) { ElMessage.error(err.message); }
  finally { loading.value = false; }
}

function openEditor(admin) {
  editingId.value = admin?.id || null;
  Object.assign(form, admin ? {
    username: admin.username, email: admin.email || "", password: "", displayName: admin.displayName,
    role: admin.role, status: admin.status === "locked" ? "active" : admin.status,
    permissionKeys: Object.entries(admin.permissions || {}).filter(([, enabled]) => enabled).map(([key]) => key)
  } : { username: "", email: "", password: "", displayName: "", role: "operator", status: "active", permissionKeys: ["dashboard", "levels", "players"] });
  dialogOpen.value = true;
}

async function save() {
  await formRef.value.validate();
  saving.value = true;
  const permissions = Object.fromEntries(permissionOptions.map((item) => [item.value, form.permissionKeys.includes(item.value)]));
  const payload = { username: form.username, email: form.email || null, password: form.password,
    displayName: form.displayName, role: form.role, status: form.status, permissions };
  try {
    if (editingId.value) await updateAdmin(editingId.value, payload); else await createAdmin(payload);
    ElMessage.success(editingId.value ? "管理员已更新" : "管理员已创建");
    dialogOpen.value = false;
    await loadAdmins();
  } catch (err) { ElMessage.error(err.message); }
  finally { saving.value = false; }
}

async function remove(admin) {
  await ElMessageBox.confirm(`确定删除管理员“${admin.displayName || admin.username}”吗？此操作不可恢复。`, "删除管理员", { type: "warning", confirmButtonText: "确认删除", cancelButtonText: "取消" });
  try { await deleteAdmin(admin.id); ElMessage.success("管理员已删除"); await loadAdmins(); }
  catch (err) { ElMessage.error(err.message); }
}

onMounted(loadAdmins);
</script>

<template>
  <div class="page-heading">
    <div><h1>管理员管理</h1><p>创建后台账号、分配角色与权限，并管理账号状态。</p></div>
    <el-button type="primary" :icon="Plus" @click="openEditor()">新增管理员</el-button>
  </div>
  <el-alert v-if="currentUser?.role !== 'owner'" title="仅超级管理员可以查看和管理后台账号" type="warning" show-icon :closable="false" />
  <section v-else class="content-section table-section">
    <div class="table-toolbar"><strong>后台账号</strong><el-tag effect="plain">{{ admins.length }} 个</el-tag><div class="toolbar-spacer" /><el-button :icon="Refresh" :loading="loading" @click="loadAdmins">刷新</el-button></div>
    <el-table v-loading="loading" :data="admins" stripe>
      <el-table-column label="管理员" min-width="190"><template #default="scope"><div class="user-cell"><el-avatar :size="36">{{ scope.row.displayName?.slice(0,1) || scope.row.username.slice(0,1) }}</el-avatar><div><strong>{{ scope.row.displayName || scope.row.username }}</strong><span>{{ scope.row.username }}</span></div></div></template></el-table-column>
      <el-table-column prop="email" label="邮箱" min-width="190"><template #default="scope">{{ scope.row.email || '-' }}</template></el-table-column>
      <el-table-column label="角色" width="130"><template #default="scope"><el-tag :type="scope.row.role === 'owner' ? 'danger' : scope.row.role === 'operator' ? 'primary' : 'info'" effect="plain">{{ roles[scope.row.role] }}</el-tag></template></el-table-column>
      <el-table-column label="状态" width="100"><template #default="scope"><el-tag :type="scope.row.status === 'active' ? 'success' : 'warning'">{{ statuses[scope.row.status] }}</el-tag></template></el-table-column>
      <el-table-column label="最近登录" min-width="165"><template #default="scope">{{ formatDate(scope.row.lastLoginAt) }}</template></el-table-column>
      <el-table-column label="登录 IP" min-width="130"><template #default="scope">{{ scope.row.lastLoginIp || '-' }}</template></el-table-column>
      <el-table-column label="操作" width="150" fixed="right" align="right"><template #default="scope"><el-button link type="primary" :icon="Edit" @click="openEditor(scope.row)">编辑</el-button><el-button link type="danger" :icon="Delete" :disabled="scope.row.id === currentAdmin?.id" @click="remove(scope.row)">删除</el-button></template></el-table-column>
      <template #empty><el-empty description="暂无管理员" /></template>
    </el-table>
  </section>

  <el-dialog v-model="dialogOpen" :title="editingId ? '编辑管理员' : '新增管理员'" width="min(600px, 92vw)" destroy-on-close>
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
      <div class="form-grid two-columns admin-form-grid">
        <el-form-item label="用户名" prop="username"><el-input v-model.trim="form.username" autocomplete="off" /></el-form-item>
        <el-form-item label="显示名称" prop="displayName"><el-input v-model.trim="form.displayName" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model.trim="form.email" type="email" /></el-form-item>
        <el-form-item label="角色" prop="role"><el-select v-model="form.role"><el-option v-for="(label,key) in roles" :key="key" :label="label" :value="key" /></el-select></el-form-item>
        <el-form-item :label="editingId ? '重置密码（留空不修改）' : '初始密码'" prop="password"><el-input v-model="form.password" type="password" show-password autocomplete="new-password" /></el-form-item>
        <el-form-item label="账号状态"><el-radio-group v-model="form.status"><el-radio-button value="active">正常</el-radio-button><el-radio-button value="disabled">禁用</el-radio-button></el-radio-group></el-form-item>
      </div>
      <el-form-item label="功能权限"><el-checkbox-group v-model="form.permissionKeys"><el-checkbox v-for="item in permissionOptions" :key="item.value" :value="item.value">{{ item.label }}</el-checkbox></el-checkbox-group><div class="field-help">超级管理员始终拥有全部系统权限；细粒度权限将在后续菜单与接口授权中使用。</div></el-form-item>
    </el-form>
    <template #footer><el-button @click="dialogOpen=false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
  </el-dialog>
</template>
