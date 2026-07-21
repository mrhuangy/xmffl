<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Delete, Edit, Plus, Refresh, Search } from "@element-plus/icons-vue";
import { createSystemControl, deleteSystemControl, fetchSystemControls, updateSystemControl } from "../api/client";
import { getUser } from "../auth/session";

const controls = ref([]);
const loading = ref(false);
const saving = ref(false);
const dialogOpen = ref(false);
const editingId = ref(null);
const formRef = ref();
const keyword = ref("");
const group = ref("");
const currentUser = getUser();
const form = reactive({});
const typeLabels = { bool: "布尔", int: "整数", decimal: "小数", string: "文本", json: "JSON" };
const filtered = computed(() => controls.value.filter((item) => {
  const search = keyword.value.trim().toLowerCase();
  return (!search || item.controlKey.toLowerCase().includes(search) || item.name.toLowerCase().includes(search)) && (!group.value || item.controlGroup === group.value);
}));
const groups = computed(() => [...new Set(controls.value.map((item) => item.controlGroup))].sort());
const rules = {
  controlKey: [{ required: true, message: "请输入配置键", trigger: "blur" }],
  controlGroup: [{ required: true, message: "请输入分组", trigger: "blur" }],
  name: [{ required: true, message: "请输入显示名称", trigger: "blur" }]
};

function defaults() {
  return { controlKey: "", controlGroup: "general", name: "", description: "", valueType: "string",
    valueText: "", valueJsonText: "{}", defaultValueText: "", defaultValueJsonText: "{}",
    enabled: true, isPublic: false, sortOrder: 0, effectiveFrom: null, effectiveUntil: null };
}
function formatDate(value) { return value ? new Intl.DateTimeFormat("zh-CN", { dateStyle: "medium", timeStyle: "short" }).format(new Date(value)) : "长期"; }
function displayValue(item) {
  if (item.valueType === "json") return JSON.stringify(item.valueJson);
  if (item.valueType === "bool") return item.valueText === "true" ? "开启" : "关闭";
  return item.valueText ?? "";
}

async function load() {
  loading.value = true;
  try { controls.value = (await fetchSystemControls()).controls || []; }
  catch (err) { ElMessage.error(err.message); }
  finally { loading.value = false; }
}

function openEditor(item) {
  editingId.value = item?.id || null;
  Object.keys(form).forEach((key) => delete form[key]);
  Object.assign(form, item ? {
    controlKey: item.controlKey, controlGroup: item.controlGroup, name: item.name,
    description: item.description, valueType: item.valueType, valueText: item.valueText ?? "",
    valueJsonText: JSON.stringify(item.valueJson ?? {}, null, 2), defaultValueText: item.defaultValueText ?? "",
    defaultValueJsonText: JSON.stringify(item.defaultValueJson ?? {}, null, 2), enabled: item.enabled,
    isPublic: item.isPublic, sortOrder: item.sortOrder, effectiveFrom: item.effectiveFrom || null,
    effectiveUntil: item.effectiveUntil || null
  } : defaults());
  dialogOpen.value = true;
}

async function save() {
  await formRef.value.validate();
  let valueJson = null;
  let defaultValueJson = null;
  if (form.valueType === "json") {
    try { valueJson = JSON.parse(form.valueJsonText); }
    catch { ElMessage.warning("当前值不是有效 JSON"); return; }
    if (form.defaultValueJsonText.trim()) {
      try { defaultValueJson = JSON.parse(form.defaultValueJsonText); }
      catch { ElMessage.warning("默认值不是有效 JSON"); return; }
    }
  }
  if (["int", "decimal"].includes(form.valueType) && form.valueText === "") { ElMessage.warning("请输入配置值"); return; }
  saving.value = true;
  const payload = { controlKey: form.controlKey, controlGroup: form.controlGroup, name: form.name,
    description: form.description, valueType: form.valueType,
    valueText: form.valueType === "json" ? null : String(form.valueText), valueJson,
    defaultValueText: form.valueType === "json" ? null : (form.defaultValueText === "" ? null : String(form.defaultValueText)),
    defaultValueJson, enabled: form.enabled, isPublic: form.isPublic, sortOrder: form.sortOrder,
    effectiveFrom: form.effectiveFrom || null, effectiveUntil: form.effectiveUntil || null };
  try {
    if (editingId.value) await updateSystemControl(editingId.value, payload); else await createSystemControl(payload);
    ElMessage.success(editingId.value ? "特殊配置已更新" : "特殊配置已创建"); dialogOpen.value = false; await load();
  } catch (err) { ElMessage.error(err.message); }
  finally { saving.value = false; }
}

async function remove(item) {
  await ElMessageBox.confirm(`确定删除配置“${item.name}（${item.controlKey}）”吗？删除后客户端将无法获取该配置。`, "删除特殊配置", { type: "warning", confirmButtonText: "确认删除", cancelButtonText: "取消" });
  try { await deleteSystemControl(item.id); ElMessage.success("配置已删除"); await load(); }
  catch (err) { ElMessage.error(err.message); }
}

onMounted(load);
</script>

<template>
  <div class="page-heading"><div><h1>特殊配置</h1><p>管理全局开关、客户端策略、玩法功能和运营公告。</p></div><el-button v-if="currentUser?.role === 'owner'" type="primary" :icon="Plus" @click="openEditor()">新增配置</el-button></div>
  <el-alert v-if="currentUser?.role !== 'owner'" title="仅超级管理员可以管理特殊配置" type="warning" show-icon :closable="false" />
  <section v-else class="content-section table-section">
    <div class="table-toolbar"><el-input v-model="keyword" :prefix-icon="Search" clearable placeholder="搜索配置名称或键" class="search-input" /><el-select v-model="group" clearable placeholder="全部分组" class="status-select"><el-option v-for="item in groups" :key="item" :label="item" :value="item" /></el-select><div class="toolbar-spacer" /><el-button :icon="Refresh" :loading="loading" @click="load">刷新</el-button></div>
    <el-table v-loading="loading" :data="filtered" stripe height="calc(100vh - 220px)">
      <el-table-column label="配置项" min-width="240"><template #default="scope"><div class="control-name"><strong>{{ scope.row.name }}</strong><span>{{ scope.row.controlKey }}</span></div></template></el-table-column>
      <el-table-column prop="controlGroup" label="分组" width="100"><template #default="scope"><el-tag effect="plain">{{ scope.row.controlGroup }}</el-tag></template></el-table-column>
      <el-table-column label="当前值" min-width="190"><template #default="scope"><span class="control-value" :title="displayValue(scope.row)">{{ displayValue(scope.row) }}</span></template></el-table-column>
      <el-table-column label="类型" width="80"><template #default="scope">{{ typeLabels[scope.row.valueType] }}</template></el-table-column>
      <el-table-column label="下发" width="75"><template #default="scope"><el-tag :type="scope.row.isPublic ? 'success' : 'info'" size="small">{{ scope.row.isPublic ? '公开' : '内部' }}</el-tag></template></el-table-column>
      <el-table-column label="状态" width="75"><template #default="scope"><el-tag :type="scope.row.enabled ? 'success' : 'info'" size="small">{{ scope.row.enabled ? '启用' : '停用' }}</el-tag></template></el-table-column>
      <el-table-column label="版本" width="70"><template #default="scope">v{{ scope.row.version }}</template></el-table-column>
      <el-table-column label="生效截止" min-width="145"><template #default="scope">{{ formatDate(scope.row.effectiveUntil) }}</template></el-table-column>
      <el-table-column label="操作" width="145" fixed="right" align="right"><template #default="scope"><el-button link type="primary" :icon="Edit" @click="openEditor(scope.row)">编辑</el-button><el-button link type="danger" :icon="Delete" @click="remove(scope.row)">删除</el-button></template></el-table-column>
    </el-table>
  </section>

  <el-dialog v-model="dialogOpen" :title="editingId ? '编辑特殊配置' : '新增特殊配置'" width="min(720px, 94vw)" destroy-on-close>
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
      <div class="form-grid two-columns admin-form-grid">
        <el-form-item label="配置键" prop="controlKey"><el-input v-model.trim="form.controlKey" placeholder="例如 game.feature_enabled" /></el-form-item>
        <el-form-item label="分组" prop="controlGroup"><el-input v-model.trim="form.controlGroup" placeholder="system / client / game / notice" /></el-form-item>
        <el-form-item label="显示名称" prop="name"><el-input v-model.trim="form.name" /></el-form-item>
        <el-form-item label="值类型"><el-select v-model="form.valueType"><el-option v-for="(label,key) in typeLabels" :key="key" :label="label" :value="key" /></el-select></el-form-item>
      </div>
      <el-form-item label="说明"><el-input v-model="form.description" type="textarea" :rows="2" /></el-form-item>
      <template v-if="form.valueType === 'bool'"><div class="form-grid two-columns admin-form-grid"><el-form-item label="当前值"><el-switch v-model="form.valueText" active-value="true" inactive-value="false" active-text="开启" inactive-text="关闭" /></el-form-item><el-form-item label="默认值"><el-switch v-model="form.defaultValueText" active-value="true" inactive-value="false" /></el-form-item></div></template>
      <template v-else-if="form.valueType === 'json'"><el-form-item label="当前 JSON"><el-input v-model="form.valueJsonText" type="textarea" :rows="6" class="json-editor" /></el-form-item><el-form-item label="默认 JSON"><el-input v-model="form.defaultValueJsonText" type="textarea" :rows="4" class="json-editor" /></el-form-item></template>
      <template v-else><div class="form-grid two-columns admin-form-grid"><el-form-item label="当前值"><el-input v-model="form.valueText" :type="['int','decimal'].includes(form.valueType) ? 'number' : 'text'" /></el-form-item><el-form-item label="默认值"><el-input v-model="form.defaultValueText" :type="['int','decimal'].includes(form.valueType) ? 'number' : 'text'" /></el-form-item></div></template>
      <div class="form-grid two-columns admin-form-grid"><el-form-item label="排序"><el-input-number v-model="form.sortOrder" /></el-form-item><el-form-item label="可用状态"><el-checkbox v-model="form.enabled">启用</el-checkbox><el-checkbox v-model="form.isPublic">允许下发客户端</el-checkbox></el-form-item><el-form-item label="生效时间"><el-date-picker v-model="form.effectiveFrom" type="datetime" value-format="YYYY-MM-DDTHH:mm:ss+08:00" placeholder="立即生效" /></el-form-item><el-form-item label="失效时间"><el-date-picker v-model="form.effectiveUntil" type="datetime" value-format="YYYY-MM-DDTHH:mm:ss+08:00" placeholder="长期有效" /></el-form-item></div>
    </el-form>
    <template #footer><el-button @click="dialogOpen=false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
  </el-dialog>
</template>
