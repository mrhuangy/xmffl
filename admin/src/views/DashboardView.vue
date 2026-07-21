<script setup>
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { ArrowRight, CircleCheck, Collection, VideoPlay } from "@element-plus/icons-vue";
import { fetchAdConfig, fetchLevels } from "../api/client";

const router = useRouter();
const loading = ref(true);
const error = ref("");
const levels = ref([]);
const ads = ref(null);
const enabledLevels = computed(() => levels.value.filter((item) => item.enabled).length);
const modes = computed(() => new Set(levels.value.map((item) => item.mode)).size);

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const [levelData, adData] = await Promise.all([fetchLevels(), fetchAdConfig()]);
    levels.value = levelData.levels || [];
    ads.value = adData;
  } catch (err) {
    error.value = err.message;
  } finally {
    loading.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="page-heading">
    <div><h1>运营概览</h1><p>关卡内容与广告策略的当前发布状态。</p></div>
    <el-button :loading="loading" @click="load">刷新数据</el-button>
  </div>
  <el-alert v-if="error" :title="error" type="error" show-icon :closable="false" />

  <el-skeleton :loading="loading" animated :rows="5">
    <div class="metric-grid">
      <section class="metric-item">
        <el-icon class="metric-icon green"><Collection /></el-icon>
        <div><span>关卡总数</span><strong>{{ levels.length }}</strong><small>{{ enabledLevels }} 个已启用</small></div>
      </section>
      <section class="metric-item">
        <el-icon class="metric-icon blue"><CircleCheck /></el-icon>
        <div><span>玩法模式</span><strong>{{ modes }}</strong><small>当前配置覆盖</small></div>
      </section>
      <section class="metric-item">
        <el-icon class="metric-icon amber"><VideoPlay /></el-icon>
        <div><span>每日插屏上限</span><strong>{{ ads?.maxInterstitialPerDay ?? '-' }}</strong><small>配置版本 v{{ ads?.version ?? '-' }}</small></div>
      </section>
    </div>

    <section class="content-section">
      <div class="section-title"><div><h2>配置入口</h2><p>修改后将直接影响小游戏远程配置。</p></div></div>
      <el-table :data="[
        { name: '关卡配置', detail: `${levels.length} 个关卡，${enabledLevels} 个启用`, path: '/levels' },
        { name: '广告频控', detail: `每 ${ads?.interstitialEveryLevels ?? '-'} 关允许插屏`, path: '/ads' }
      ]">
        <el-table-column prop="name" label="配置模块" min-width="180" />
        <el-table-column prop="detail" label="当前状态" min-width="260" />
        <el-table-column label="操作" width="100" align="right">
          <template #default="scope"><el-button link type="primary" @click="router.push(scope.row.path)">管理 <el-icon><ArrowRight /></el-icon></el-button></template>
        </el-table-column>
      </el-table>
    </section>
  </el-skeleton>
</template>
