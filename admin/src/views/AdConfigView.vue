<script setup>
import { onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import { Check, Refresh } from "@element-plus/icons-vue";
import { fetchAdConfig, saveAdConfig } from "../api/client";

const scenes = [
  { label: "首页", value: "home" },
  { label: "结算页", value: "result" },
  { label: "关卡选择", value: "level_select" }
];
const formRef = ref();
const loading = ref(false);
const saving = ref(false);
const form = reactive({ noInterstitialBeforeLevel: 4, interstitialEveryLevels: 4, maxInterstitialPerDay: 10, maxRevivePerLevel: 1, bannerEnabledScenes: [], version: 1 });
const rules = {
  interstitialEveryLevels: [{ required: true, message: "间隔必须大于 0", trigger: "change" }]
};

async function loadConfig() {
  loading.value = true;
  try {
    Object.assign(form, await fetchAdConfig());
  } catch (err) {
    ElMessage.error(err.message);
  } finally {
    loading.value = false;
  }
}

async function persistConfig() {
  await formRef.value.validate();
  saving.value = true;
  try {
    await saveAdConfig(form);
    ElMessage.success("广告频控已保存");
    await loadConfig();
  } catch (err) {
    ElMessage.error(err.message);
  } finally {
    saving.value = false;
  }
}

onMounted(loadConfig);
</script>

<template>
  <div class="page-heading">
    <div><h1>广告频控</h1><p>控制插屏出现节奏、单局复活次数和 Banner 展示场景。</p></div>
    <div class="heading-actions"><el-button :icon="Refresh" :loading="loading" @click="loadConfig">重载</el-button><el-button type="primary" :icon="Check" :loading="saving" @click="persistConfig">保存配置</el-button></div>
  </div>

  <div v-loading="loading" class="config-layout">
    <section class="content-section config-main">
      <div class="section-title"><div><h2>插屏广告</h2><p>只在关卡之间的自然间隙触发。</p></div><el-tag effect="plain">配置版本 v{{ form.version }}</el-tag></div>
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <div class="form-grid two-columns">
          <el-form-item label="前 N 关不展示插屏"><el-input-number v-model="form.noInterstitialBeforeLevel" :min="0" /><div class="field-help">保护新手体验，建议不少于 3 关。</div></el-form-item>
          <el-form-item label="每 N 关允许一次插屏" prop="interstitialEveryLevels"><el-input-number v-model="form.interstitialEveryLevels" :min="1" /><div class="field-help">达到间隔不代表强制展示。</div></el-form-item>
          <el-form-item label="单用户每日插屏上限"><el-input-number v-model="form.maxInterstitialPerDay" :min="0" /><div class="field-help">设为 0 将禁用每日插屏。</div></el-form-item>
          <el-form-item label="单关激励视频复活上限"><el-input-number v-model="form.maxRevivePerLevel" :min="0" /><div class="field-help">避免无限续关破坏难度。</div></el-form-item>
        </div>
      </el-form>
    </section>

    <section class="content-section config-side">
      <div class="section-title"><div><h2>Banner 场景</h2><p>局内游戏页默认不开放。</p></div></div>
      <el-checkbox-group v-model="form.bannerEnabledScenes" class="scene-list">
        <label v-for="scene in scenes" :key="scene.value" class="scene-option">
          <div><strong>{{ scene.label }}</strong><span>{{ scene.value }}</span></div>
          <el-checkbox :value="scene.value" />
        </label>
      </el-checkbox-group>
      <el-alert title="Banner 不得覆盖牌阵和核心操作按钮" type="warning" :closable="false" show-icon />
    </section>
  </div>
</template>
