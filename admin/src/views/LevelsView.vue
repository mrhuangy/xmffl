<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import { Edit, Plus, Refresh, Search } from "@element-plus/icons-vue";
import { fetchLevels, saveLevel } from "../api/client";

const levels = ref([]);
const loading = ref(false);
const saving = ref(false);
const drawerOpen = ref(false);
const formRef = ref();
const keyword = ref("");
const status = ref("all");
const page = ref(1);
const pageSize = ref(15);
const form = reactive({});

const modeLabels = { normal: "普通", time_limit: "限时", step_limit: "限步" };
const filtered = computed(() => levels.value.filter((level) => {
  const matchesKeyword = !keyword.value || String(level.levelId).includes(keyword.value) || level.themeId.includes(keyword.value.trim());
  const matchesStatus = status.value === "all" || level.enabled === (status.value === "enabled");
  return matchesKeyword && matchesStatus;
}));
const pagedLevels = computed(() => filtered.value.slice((page.value - 1) * pageSize.value, page.value * pageSize.value));
const drawerTitle = computed(() => levels.value.some((item) => item.levelId === form.levelId) ? `编辑第 ${form.levelId} 关` : "新增关卡");

const rules = {
  levelId: [{ required: true, message: "请输入关卡编号", trigger: "blur" }],
  themeId: [{ required: true, message: "请输入主题 ID", trigger: "blur" }],
  rows: [{ required: true, message: "请输入行数", trigger: "blur" }],
  cols: [{ required: true, message: "请输入列数", trigger: "blur" }],
  pairCount: [{ required: true, message: "请输入配对数", trigger: "blur" }]
};

function defaults() {
  return {
    levelId: Math.max(0, ...levels.value.map((item) => item.levelId)) + 1,
    rows: 4, cols: 4, pairCount: 8, mode: "normal", themeId: "animal",
    initialPreviewMs: 2000, flipBackDelayMs: 700, levelTimeLimitSeconds: 120,
    maxMismatchCount: 12, showSteps: true, showTimer: true, showMismatch: true,
    hintHighlightMs: 1300, coinRewardBase: 10, staminaCost: 1,
    excellentStepThreshold: 12, normalStepThreshold: 18,
    excellentTimeThreshold: 70, normalTimeThreshold: 105,
    timeLimitSeconds: null, stepLimit: null, enabled: true, version: 1
  };
}

async function loadLevels() {
  loading.value = true;
  try {
    const data = await fetchLevels();
    levels.value = data.levels || [];
  } catch (err) {
    ElMessage.error(err.message);
  } finally {
    loading.value = false;
  }
}

function openEditor(level) {
  Object.keys(form).forEach((key) => delete form[key]);
  Object.assign(form, level ? structuredClone(level) : defaults());
  drawerOpen.value = true;
}

async function persistLevel() {
  await formRef.value.validate();
  const slots = form.rows * form.cols;
  const cards = form.pairCount * 2;
  if (slots !== cards && !(slots === cards + 1 && slots % 2 === 1)) {
    ElMessage.warning("牌阵格数必须等于牌数，奇数牌阵仅允许中心留空一格");
    return;
  }
  if (form.excellentStepThreshold > form.normalStepThreshold) {
    ElMessage.warning("三星步数阈值不能高于二星阈值");
    return;
  }
  saving.value = true;
  try {
    await saveLevel(form);
    ElMessage.success(`第 ${form.levelId} 关已保存`);
    drawerOpen.value = false;
    await loadLevels();
  } catch (err) {
    ElMessage.error(err.message);
  } finally {
    saving.value = false;
  }
}

onMounted(loadLevels);
</script>

<template>
  <div class="page-heading">
    <div><h1>关卡管理</h1><p>维护牌阵、难度、奖励和局内展示配置。</p></div>
    <el-button type="primary" :icon="Plus" @click="openEditor()">新增关卡</el-button>
  </div>

  <section class="content-section table-section">
    <div class="table-toolbar">
      <el-input v-model="keyword" :prefix-icon="Search" clearable placeholder="搜索关卡编号或主题" class="search-input" @input="page = 1" />
      <el-segmented v-model="status" :options="[
        { label: '全部', value: 'all' }, { label: '已启用', value: 'enabled' }, { label: '已停用', value: 'disabled' }
      ]" @change="page = 1" />
      <div class="toolbar-spacer" />
      <el-button :icon="Refresh" :loading="loading" @click="loadLevels">刷新</el-button>
    </div>

    <el-table v-loading="loading" :data="pagedLevels" height="calc(100vh - 255px)" stripe>
      <el-table-column prop="levelId" label="关卡" width="90"><template #default="scope"><strong>#{{ scope.row.levelId }}</strong></template></el-table-column>
      <el-table-column label="牌阵" width="120"><template #default="scope">{{ scope.row.rows }} x {{ scope.row.cols }}</template></el-table-column>
      <el-table-column prop="pairCount" label="配对" width="90"><template #default="scope">{{ scope.row.pairCount }} 对</template></el-table-column>
      <el-table-column prop="mode" label="模式" width="100"><template #default="scope"><el-tag effect="plain">{{ modeLabels[scope.row.mode] }}</el-tag></template></el-table-column>
      <el-table-column prop="themeId" label="主题" min-width="120" />
      <el-table-column label="时长 / 错配" min-width="150"><template #default="scope">{{ scope.row.levelTimeLimitSeconds }} 秒 / {{ scope.row.maxMismatchCount }} 次</template></el-table-column>
      <el-table-column prop="version" label="版本" width="80"><template #default="scope">v{{ scope.row.version }}</template></el-table-column>
      <el-table-column label="状态" width="90"><template #default="scope"><el-tag :type="scope.row.enabled ? 'success' : 'info'" effect="light">{{ scope.row.enabled ? '启用' : '停用' }}</el-tag></template></el-table-column>
      <el-table-column label="操作" width="90" fixed="right" align="right"><template #default="scope"><el-button link type="primary" :icon="Edit" @click="openEditor(scope.row)">编辑</el-button></template></el-table-column>
      <template #empty><el-empty description="没有符合条件的关卡" /></template>
    </el-table>
    <div class="pagination-row"><span>共 {{ filtered.length }} 个关卡</span><el-pagination v-model:current-page="page" v-model:page-size="pageSize" layout="prev, pager, next" :total="filtered.length" /></div>
  </section>

  <el-drawer v-model="drawerOpen" :title="drawerTitle" size="min(720px, 92vw)" destroy-on-close>
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top" class="drawer-form">
      <div class="form-section-title"><h3>基础配置</h3><el-switch v-model="form.enabled" active-text="启用关卡" /></div>
      <div class="form-grid">
        <el-form-item label="关卡编号" prop="levelId"><el-input-number v-model="form.levelId" :min="1" :controls="false" /></el-form-item>
        <el-form-item label="主题 ID" prop="themeId"><el-input v-model.trim="form.themeId" /></el-form-item>
        <el-form-item label="模式"><el-select v-model="form.mode"><el-option label="普通模式" value="normal" /><el-option label="限时模式" value="time_limit" /><el-option label="限步模式" value="step_limit" /></el-select></el-form-item>
        <el-form-item label="行数" prop="rows"><el-input-number v-model="form.rows" :min="1" :max="10" /></el-form-item>
        <el-form-item label="列数" prop="cols"><el-input-number v-model="form.cols" :min="1" :max="10" /></el-form-item>
        <el-form-item label="配对数" prop="pairCount"><el-input-number v-model="form.pairCount" :min="1" :max="50" /></el-form-item>
      </div>
      <div class="form-section-title"><h3>游戏节奏</h3></div>
      <div class="form-grid">
        <el-form-item label="开局预览 (ms)"><el-input-number v-model="form.initialPreviewMs" :min="0" :step="100" /></el-form-item>
        <el-form-item label="错配翻回延迟 (ms)"><el-input-number v-model="form.flipBackDelayMs" :min="0" :step="100" /></el-form-item>
        <el-form-item label="关卡时长上限 (秒)"><el-input-number v-model="form.levelTimeLimitSeconds" :min="1" /></el-form-item>
        <el-form-item label="错配次数上限"><el-input-number v-model="form.maxMismatchCount" :min="0" /></el-form-item>
        <el-form-item label="提示高亮 (ms)"><el-input-number v-model="form.hintHighlightMs" :min="100" :step="100" /></el-form-item>
        <el-form-item label="体力消耗"><el-input-number v-model="form.staminaCost" :min="0" /></el-form-item>
      </div>
      <div class="form-section-title"><h3>星级与奖励</h3></div>
      <div class="form-grid">
        <el-form-item label="三星步数阈值"><el-input-number v-model="form.excellentStepThreshold" :min="1" /></el-form-item>
        <el-form-item label="二星步数阈值"><el-input-number v-model="form.normalStepThreshold" :min="1" /></el-form-item>
        <el-form-item label="基础金币奖励"><el-input-number v-model="form.coinRewardBase" :min="0" /></el-form-item>
        <el-form-item label="三星时间阈值 (秒)"><el-input-number v-model="form.excellentTimeThreshold" :min="1" /></el-form-item>
        <el-form-item label="二星时间阈值 (秒)"><el-input-number v-model="form.normalTimeThreshold" :min="1" /></el-form-item>
      </div>
      <div class="form-section-title"><h3>局内显示</h3></div>
      <el-checkbox v-model="form.showSteps">显示步数</el-checkbox><el-checkbox v-model="form.showTimer">显示计时</el-checkbox><el-checkbox v-model="form.showMismatch">显示错配次数</el-checkbox>
    </el-form>
    <template #footer><div class="drawer-actions"><el-button @click="drawerOpen = false">取消</el-button><el-button type="primary" :loading="saving" @click="persistLevel">保存配置</el-button></div></template>
  </el-drawer>
</template>
