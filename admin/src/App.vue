<script setup>
import { ref } from "vue";
import AdConfigView from "./views/AdConfigView.vue";
import LevelsView from "./views/LevelsView.vue";

const activeView = ref("levels");
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div class="brand">
        <span class="brand-mark">FP</span>
        <div>
          <strong>翻牌消除</strong>
          <small>运营后台</small>
        </div>
      </div>

      <nav class="nav-list">
        <button :class="{ active: activeView === 'levels' }" @click="activeView = 'levels'">
          关卡配置
        </button>
        <button :class="{ active: activeView === 'ads' }" @click="activeView = 'ads'">
          广告频控
        </button>
      </nav>
    </aside>

    <main class="main-area">
      <LevelsView v-if="activeView === 'levels'" />
      <AdConfigView v-else />
    </main>
  </div>
</template>
