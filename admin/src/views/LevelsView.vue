<script setup>
import { computed, onMounted, ref } from "vue";
import { fetchLevels, saveLevel } from "../api/client";

const levels = ref([]);
const selectedLevelId = ref(null);
const loading = ref(false);
const saving = ref(false);
const error = ref("");
const message = ref("");

const selectedLevel = computed(() => levels.value.find((level) => level.levelId === selectedLevelId.value));

async function loadLevels() {
  loading.value = true;
  error.value = "";
  try {
    const data = await fetchLevels();
    levels.value = data.levels || [];
    selectedLevelId.value = levels.value[0]?.levelId ?? null;
  } catch (err) {
    error.value = err.message;
  } finally {
    loading.value = false;
  }
}

async function persistLevel() {
  if (!selectedLevel.value) return;
  saving.value = true;
  error.value = "";
  message.value = "";
  try {
    await saveLevel(selectedLevel.value);
    message.value = `第 ${selectedLevel.value.levelId} 关已保存`;
  } catch (err) {
    error.value = err.message;
  } finally {
    saving.value = false;
  }
}

function addLevel() {
  const nextId = Math.max(0, ...levels.value.map((level) => level.levelId)) + 1;
  levels.value.push({
    levelId: nextId,
    rows: 4,
    cols: 4,
    pairCount: 8,
    mode: "normal",
    themeId: "animal",
    initialPreviewMs: 2000,
    flipBackDelayMs: 700,
    levelTimeLimitSeconds: 120,
    maxMismatchCount: 12,
    excellentStepThreshold: 12,
    normalStepThreshold: 18,
    excellentTimeThreshold: 70,
    normalTimeThreshold: 105,
    enabled: true
  });
  selectedLevelId.value = nextId;
}

onMounted(loadLevels);
</script>

<template>
  <section class="page-header">
    <div>
      <h1>关卡配置</h1>
      <p>维护小游戏远程关卡参数，字段对齐 LevelConfig。</p>
    </div>
    <button class="primary-button" @click="addLevel">新增关卡</button>
  </section>

  <div v-if="error" class="notice error">{{ error }}</div>
  <div v-if="message" class="notice success">{{ message }}</div>

  <div class="workbench">
    <section class="list-panel">
      <div class="panel-title">
        <strong>关卡列表</strong>
        <span>{{ levels.length }} 个</span>
      </div>
      <button
        v-for="level in levels"
        :key="level.levelId"
        class="level-row"
        :class="{ active: level.levelId === selectedLevelId }"
        @click="selectedLevelId = level.levelId"
      >
        <span>第 {{ level.levelId }} 关</span>
        <small>{{ level.rows }}x{{ level.cols }} / {{ level.pairCount }} 对</small>
      </button>
      <div v-if="loading" class="empty-state">加载中</div>
    </section>

    <section v-if="selectedLevel" class="editor-panel">
      <div class="panel-title">
        <strong>第 {{ selectedLevel.levelId }} 关</strong>
        <label class="switch">
          <input v-model="selectedLevel.enabled" type="checkbox" />
          启用
        </label>
      </div>

      <div class="form-grid">
        <label>
          行数
          <input v-model.number="selectedLevel.rows" type="number" min="1" />
        </label>
        <label>
          列数
          <input v-model.number="selectedLevel.cols" type="number" min="1" />
        </label>
        <label>
          配对数
          <input v-model.number="selectedLevel.pairCount" type="number" min="1" />
        </label>
        <label>
          模式
          <select v-model="selectedLevel.mode">
            <option value="normal">普通模式</option>
            <option value="time_limit">限时模式</option>
            <option value="step_limit">限步模式</option>
          </select>
        </label>
        <label>
          主题 ID
          <input v-model.trim="selectedLevel.themeId" />
        </label>
        <label>
          预览毫秒
          <input v-model.number="selectedLevel.initialPreviewMs" type="number" min="0" />
        </label>
        <label>
          翻回延迟
          <input v-model.number="selectedLevel.flipBackDelayMs" type="number" min="0" />
        </label>
        <label>
          时长上限秒
          <input v-model.number="selectedLevel.levelTimeLimitSeconds" type="number" min="1" />
        </label>
        <label>
          错配上限
          <input v-model.number="selectedLevel.maxMismatchCount" type="number" min="0" />
        </label>
        <label>
          3 星步数
          <input v-model.number="selectedLevel.excellentStepThreshold" type="number" min="1" />
        </label>
        <label>
          2 星步数
          <input v-model.number="selectedLevel.normalStepThreshold" type="number" min="1" />
        </label>
        <label>
          3 星时间秒
          <input v-model.number="selectedLevel.excellentTimeThreshold" type="number" min="1" />
        </label>
        <label>
          2 星时间秒
          <input v-model.number="selectedLevel.normalTimeThreshold" type="number" min="1" />
        </label>
      </div>

      <footer class="editor-actions">
        <span>牌面总数必须等于配对数 x 2。</span>
        <button class="primary-button" :disabled="saving" @click="persistLevel">
          {{ saving ? "保存中" : "保存" }}
        </button>
      </footer>
    </section>
  </div>
</template>
