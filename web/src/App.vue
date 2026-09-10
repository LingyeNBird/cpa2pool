<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { PhUsers, PhSlidersHorizontal, PhReceipt, PhChartLine } from '@phosphor-icons/vue';
import { authenticated, login, restoreLogin, logout, act, failure, busy } from './api';
import ParticipantsPage from './components/ParticipantsPage.vue';
import PricesPage from './components/PricesPage.vue';
import ReportsPage from './components/ReportsPage.vue';
import PageActions from './components/PageActions.vue';
import TooltipIcon from './components/components/TooltipIcon.vue';
import './App.css';
const credential = ref('');
const restoring = ref(true);
const section = ref('participants');
const revision = ref(0);
const tabs = [
  { id: 'participants', title: '参与者', icon: PhUsers },
  { id: 'prices', title: '计费设置', icon: PhSlidersHorizontal },
  { id: 'reports', title: '账单统计', icon: PhReceipt },
];
async function signIn() {
  await act(async () => {
    await login(credential.value);
    credential.value = '';
  });
}
onMounted(async () => {
  try {
    await act(restoreLogin);
  } finally {
    restoring.value = false;
  }
});
</script>
<template>
  <main class="shell">
    <header class="app-header">
      <div class="brand">
        <PhChartLine :size="27" weight="bold" />
        <h1>拼车额度</h1>
      </div>
      <PageActions v-if="authenticated" :busy="busy" @refresh="revision++" @logout="logout" />
    </header>
    <div v-if="failure" class="alert alert-error error-bar" role="alert">
      <span>{{ failure }}</span
      ><button class="btn btn-sm btn-ghost" @click="failure = ''">关闭</button>
    </div>
    <div v-if="restoring" class="panel login-panel" role="status">正在连接…</div>
    <form v-else-if="!authenticated" class="panel login-panel" @submit.prevent="signIn">
      <h2>
        连接管理中心
        <TooltipIcon
          tip="自动使用同源 CPA 管理面板记住的登录。手动输入的密钥仅保存在当前页面内存中。"
          label="连接说明"
          :size="19"
        />
      </h2>
      <label class="field"
        ><span>管理密钥</span
        ><input
          v-model="credential"
          class="input"
          type="password"
          autocomplete="current-password"
          required
          autofocus
      /></label>
      <button class="btn btn-primary" :disabled="busy">{{ busy ? '连接中' : '连接' }}</button>
    </form>
    <template v-else>
      <nav class="section-tabs" aria-label="插件功能">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          class="btn"
          :class="section === tab.id ? 'btn-primary' : 'btn-ghost'"
          :aria-current="section === tab.id ? 'page' : undefined"
          @click="section = tab.id"
        >
          <component :is="tab.icon" :size="19" />{{ tab.title }}
        </button>
      </nav>
      <ParticipantsPage v-if="section === 'participants'" :key="`p${revision}`" />
      <PricesPage v-else-if="section === 'prices'" :key="`c${revision}`" />
      <ReportsPage v-else :key="`r${revision}`" />
    </template>
  </main>
</template>
