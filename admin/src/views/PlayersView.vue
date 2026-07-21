<script setup>
import { onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import { Refresh, Search, View } from "@element-plus/icons-vue";
import { fetchPlayerDetail, fetchPlayers } from "../api/client";

const loading = ref(false);
const detailLoading = ref(false);
const drawerOpen = ref(false);
const players = ref([]);
const total = ref(0);
const detail = ref(null);
const filters = reactive({ keyword: "", status: "", page: 1, pageSize: 20 });
const statusLabels = { active: "正常", blocked: "已封禁", deleted: "已删除" };
const reasonLabels = { completed: "通关", time_out: "超时", mismatch_limit: "错配超限", quit: "退出", unknown: "未知" };

function formatDate(value) {
  return value ? new Intl.DateTimeFormat("zh-CN", { dateStyle: "medium", timeStyle: "short" }).format(new Date(value)) : "-";
}
function formatDuration(ms) {
  if (!Number.isFinite(ms)) return "-";
  return `${Math.floor(ms / 60000)}:${String(Math.floor(ms / 1000) % 60).padStart(2, "0")}`;
}
function maskOpenId(value) {
  return value?.length > 14 ? `${value.slice(0, 7)}...${value.slice(-5)}` : value;
}

async function loadPlayers() {
  loading.value = true;
  try {
    const data = await fetchPlayers(filters);
    players.value = data.players || [];
    total.value = data.total || 0;
  } catch (err) {
    ElMessage.error(err.message);
  } finally {
    loading.value = false;
  }
}

function search() { filters.page = 1; loadPlayers(); }

async function openDetail(player) {
  drawerOpen.value = true;
  detailLoading.value = true;
  detail.value = null;
  try {
    detail.value = await fetchPlayerDetail(player.id);
  } catch (err) {
    ElMessage.error(err.message);
    drawerOpen.value = false;
  } finally {
    detailLoading.value = false;
  }
}

onMounted(loadPlayers);
</script>

<template>
  <div class="page-heading">
    <div><h1>用户管理</h1><p>查看玩家账号、游戏进度、资源余额和最近对局。</p></div>
    <el-button :icon="Refresh" :loading="loading" @click="loadPlayers">刷新</el-button>
  </div>

  <section class="content-section table-section">
    <div class="table-toolbar">
      <el-input v-model="filters.keyword" :prefix-icon="Search" clearable placeholder="搜索昵称或 OpenID" class="search-input" @keyup.enter="search" @clear="search" />
      <el-select v-model="filters.status" clearable placeholder="全部状态" class="status-select" @change="search">
        <el-option label="正常" value="active" /><el-option label="已封禁" value="blocked" /><el-option label="已删除" value="deleted" />
      </el-select>
      <el-button type="primary" @click="search">查询</el-button>
      <div class="toolbar-spacer" /><span class="result-count">{{ total }} 位用户</span>
    </div>
    <el-table v-loading="loading" :data="players" height="calc(100vh - 255px)" stripe>
      <el-table-column label="用户" min-width="210">
        <template #default="scope"><div class="user-cell"><el-avatar :size="36" :src="scope.row.avatarUrl">{{ scope.row.nickname?.slice(0,1) || '游' }}</el-avatar><div><strong>{{ scope.row.nickname || '微信玩家' }}</strong><span :title="scope.row.openId">{{ maskOpenId(scope.row.openId) }}</span></div></div></template>
      </el-table-column>
      <el-table-column prop="currentLevel" label="当前关卡" width="100"><template #default="scope">第 {{ scope.row.currentLevel }} 关</template></el-table-column>
      <el-table-column label="游戏进度" min-width="140"><template #default="scope">{{ scope.row.completedCount }} 关 / {{ scope.row.totalGames }} 局</template></el-table-column>
      <el-table-column label="资源" min-width="170"><template #default="scope"><span class="resource-text">金币 {{ scope.row.coins }}</span><span class="resource-text">体力 {{ scope.row.stamina }}/{{ scope.row.maxStamina }}</span></template></el-table-column>
      <el-table-column label="最近登录" min-width="165"><template #default="scope">{{ formatDate(scope.row.lastLoginAt) }}</template></el-table-column>
      <el-table-column label="状态" width="90"><template #default="scope"><el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'" effect="light">{{ statusLabels[scope.row.status] }}</el-tag></template></el-table-column>
      <el-table-column label="操作" width="90" fixed="right" align="right"><template #default="scope"><el-button link type="primary" :icon="View" @click="openDetail(scope.row)">详情</el-button></template></el-table-column>
      <template #empty><el-empty description="暂无用户数据" /></template>
    </el-table>
    <div class="pagination-row"><span>第 {{ filters.page }} 页</span><el-pagination v-model:current-page="filters.page" v-model:page-size="filters.pageSize" layout="sizes, prev, pager, next" :page-sizes="[10,20,50,100]" :total="total" @change="loadPlayers" /></div>
  </section>

  <el-drawer v-model="drawerOpen" title="用户详细信息" size="min(780px, 94vw)" destroy-on-close>
    <div v-loading="detailLoading" class="player-detail">
      <template v-if="detail">
        <div class="detail-identity"><el-avatar :size="56" :src="detail.player.avatarUrl">{{ detail.player.nickname?.slice(0,1) || '游' }}</el-avatar><div><h2>{{ detail.player.nickname || '微信玩家' }}</h2><p>{{ detail.player.openId }}</p></div><el-tag :type="detail.player.status === 'active' ? 'success' : 'danger'">{{ statusLabels[detail.player.status] }}</el-tag></div>
        <el-descriptions :column="2" border class="detail-descriptions">
          <el-descriptions-item label="用户 ID">{{ detail.player.id }}</el-descriptions-item><el-descriptions-item label="注册时间">{{ formatDate(detail.player.createdAt) }}</el-descriptions-item>
          <el-descriptions-item label="最近登录">{{ formatDate(detail.player.lastLoginAt) }}</el-descriptions-item><el-descriptions-item label="当前关卡">第 {{ detail.progress.currentLevel }} 关</el-descriptions-item>
          <el-descriptions-item label="金币">{{ detail.progress.coins }}</el-descriptions-item><el-descriptions-item label="体力">{{ detail.progress.stamina }} / {{ detail.progress.maxStamina }}</el-descriptions-item>
          <el-descriptions-item label="提示道具">{{ detail.progress.hints }}</el-descriptions-item><el-descriptions-item label="重新预览">{{ detail.progress.previewAgainCount }}</el-descriptions-item>
          <el-descriptions-item label="消除一对">{{ detail.progress.removePairCount }}</el-descriptions-item><el-descriptions-item label="已通关">{{ detail.progress.completedLevels.length }} 关</el-descriptions-item>
        </el-descriptions>
        <div class="detail-stats"><div><span>总对局</span><strong>{{ detail.player.totalGames }}</strong></div><div><span>成功通关</span><strong>{{ detail.player.successfulGames }}</strong></div><div><span>通关率</span><strong>{{ detail.player.totalGames ? Math.round(detail.player.successfulGames / detail.player.totalGames * 100) : 0 }}%</strong></div></div>
        <div class="detail-section-title"><h3>最近对局</h3><span>最近 20 条</span></div>
        <el-table :data="detail.recentResults" max-height="330">
          <el-table-column prop="levelId" label="关卡" width="75" /><el-table-column label="结果" width="80"><template #default="scope"><el-tag :type="scope.row.success ? 'success' : 'danger'" size="small">{{ scope.row.success ? '成功' : '失败' }}</el-tag></template></el-table-column>
          <el-table-column label="原因" width="90"><template #default="scope">{{ reasonLabels[scope.row.reason] || scope.row.reason }}</template></el-table-column><el-table-column prop="stars" label="星级" width="70" /><el-table-column prop="steps" label="步数" width="70" /><el-table-column label="耗时" width="75"><template #default="scope">{{ formatDuration(scope.row.elapsedMs) }}</template></el-table-column><el-table-column label="时间" min-width="150"><template #default="scope">{{ formatDate(scope.row.completedAt) }}</template></el-table-column>
          <template #empty><el-empty description="暂无对局记录" :image-size="64" /></template>
        </el-table>
      </template>
    </div>
  </el-drawer>
</template>
