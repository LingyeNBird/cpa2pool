<script setup lang="ts">
import { ref } from 'vue';
import {
  PhUsers,
  PhSlidersHorizontal,
  PhReceipt,
  PhChartLine,
  PhQuestion,
} from '@phosphor-icons/vue';
import { authenticated, login, logout, act, failure, busy } from './api';
import ParticipantsPage from './components/ParticipantsPage.vue';
import PricesPage from './components/PricesPage.vue';
import ReportsPage from './components/ReportsPage.vue';
import PageActions from './components/PageActions.vue';
import './App.css';
const credential = ref('');
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
    <form v-if="!authenticated" class="panel login-panel" @submit.prevent="signIn">
      <h2>
        连接管理中心
        <span class="tooltip tooltip-bottom" data-tip="使用 CPA 管理密钥。仅保存在当前页面内存中。"
          ><PhQuestion :size="19" tabindex="0" aria-label="连接说明"
        /></span>
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
