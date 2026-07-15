<script setup>
import { onMounted, ref } from "vue";
import { fetchAdConfig, saveAdConfig } from "../api/client";

const scenes = ["home", "result", "level_select"];
const form = ref({
  noInterstitialBeforeLevel: 4,
  interstitialEveryLevels: 4,
  maxInterstitialPerDay: 10,
  maxRevivePerLevel: 1,
  bannerEnabledScenes: ["home", "result"]
});
const error = ref("");
const message = ref("");
const saving = ref(false);

async function loadConfig() {
  error.value = "";
  try {
    form.value = await fetchAdConfig();
  } catch (err) {
    error.value = err.message;
  }
}

async function persistConfig() {
  saving.value = true;
  error.value = "";
  message.value = "";
  try {
    await saveAdConfig(form.value);
    message.value = "广告频控已保存";
  } catch (err) {
    error.value = err.message;
  } finally {
    saving.value = false;
  }
}

onMounted(loadConfig);
</script>

<template>
  <section class="page-header">
    <div>
      <h1>广告频控</h1>
      <p>集中控制激励视频、插屏和 Banner 展示策略。</p>
    </div>
    <button class="primary-button" :disabled="saving" @click="persistConfig">
      {{ saving ? "保存中" : "保存配置" }}
    </button>
  </section>

  <div v-if="error" class="notice error">{{ error }}</div>
  <div v-if="message" class="notice success">{{ message }}</div>

  <section class="editor-panel narrow">
    <div class="form-grid">
      <label>
        前 N 关不显示插屏
        <input v-model.number="form.noInterstitialBeforeLevel" type="number" min="0" />
      </label>
      <label>
        每 N 关显示插屏
        <input v-model.number="form.interstitialEveryLevels" type="number" min="1" />
      </label>
      <label>
        每日插屏上限
        <input v-model.number="form.maxInterstitialPerDay" type="number" min="0" />
      </label>
      <label>
        单关复活上限
        <input v-model.number="form.maxRevivePerLevel" type="number" min="0" />
      </label>
    </div>

    <div class="checkbox-group">
      <strong>Banner 启用场景</strong>
      <label v-for="scene in scenes" :key="scene" class="checkbox-row">
        <input v-model="form.bannerEnabledScenes" type="checkbox" :value="scene" />
        {{ scene }}
      </label>
    </div>
  </section>
</template>
