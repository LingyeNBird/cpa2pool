<script setup lang="ts">
import AnimatedValue from './components/AnimatedValue.vue';
import ReportTable from './components/ReportTable.vue';
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import type { Participant, Bill, Period, Audit, Stat, Page, Quota } from '../types';
import { api, act, money } from '../api';
import ReportFilters from './components/ReportFilters.vue';
import SummaryStats from './components/SummaryStats.vue';
import BillsTable from './components/BillsTable.vue';
import PeriodsTable from './components/PeriodsTable.vue';
import AuditsTable from './components/AuditsTable.vue';
import SelectField from './components/SelectField.vue';
import TablePagination from './components/TablePagination.vue';
import './ReportsPage.css';
const participants = ref<Participant[]>([]);
const names = computed(() => Object.fromEntries(participants.value.map((p) => [p.id, p.name])));
const quotaNames = ref<Record<string, string>>({});
const total = ref<Stat>();
const stats = ref<Stat[]>([]);
const bills = ref<Bill[]>([]);
const history = ref<Period[]>([]);
const audits = ref<Audit[]>([]);
const tab = ref('bills');
const group = ref('model');
const offset = ref(0);
const count = ref(0);
const loading = ref(true);
const displayedTab = ref('bills');
const displayedGroup = ref('model');
const displayedOffset = ref(0);
let generation = 0;
onBeforeUnmount(() => {
  generation++;
});
let filters = new URLSearchParams();
const PAGE_SIZE = 10;
const tabs = [
  { id: 'bills', label: '请求账单' },
  { id: 'stats', label: '消费统计' },
  { id: 'periods', label: '额度周期' },
  { id: 'audits', label: '调整记录' },
];
async function load() {
  const request = ++generation;
  const requestedTab = tab.value;
  const requestedGroup = group.value;
  const requestedOffset = offset.value;
  const filterSnapshot = new URLSearchParams(filters);
  loading.value = true;
  try {
    const q = new URLSearchParams(filterSnapshot);
    q.set('limit', String(PAGE_SIZE));
    q.set('offset', String(requestedOffset));
    const summary = (await api<Stat[]>(`stats?${filterSnapshot}`))[0];
    if (request !== generation) return;
    if (requestedTab === 'bills') {
      const page = await api<Page<Bill>>(`bills?${q}`);
      if (request !== generation) return;
      bills.value = page.items;
      count.value = page.total;
    } else if (requestedTab === 'audits') {
      const page = await api<Page<Audit>>(`audits?${q}`);
      if (request !== generation) return;
      audits.value = page.items;
      count.value = page.total;
    } else if (requestedTab === 'periods') {
      const rows = await api<Period[]>(`periods?${filterSnapshot}`);
      if (request !== generation) return;
      count.value = rows.length;
      history.value = rows.slice(requestedOffset, requestedOffset + PAGE_SIZE);
    } else {
      q.set('group', requestedGroup);
      q.set('limit', '200');
      q.set('offset', '0');
      const rows = await api<Stat[]>(`stats?${q}`);
      if (request !== generation) return;
      count.value = rows.length;
      stats.value = rows.slice(requestedOffset, requestedOffset + PAGE_SIZE);
    }
    total.value = summary;
    displayedTab.value = requestedTab;
    displayedGroup.value = requestedGroup;
    displayedOffset.value = requestedOffset;
  } catch (error) {
    if (request === generation) {
      tab.value = displayedTab.value;
      group.value = displayedGroup.value;
      offset.value = displayedOffset.value;
      throw error;
    }
  } finally {
    if (request === generation) loading.value = false;
  }
}
async function apply(q: URLSearchParams) {
  filters = q;
  offset.value = 0;
  await act(load);
}
async function switchTab(value: string) {
  tab.value = value;
  offset.value = 0;
  if (value === 'periods' || value === 'audits') filters.delete('model');
  if (value === 'periods') {
    filters.delete('from');
    filters.delete('to');
  }
  await act(load);
}
async function goPage(page: number) {
  offset.value = (page - 1) * PAGE_SIZE;
  await act(load);
}
onMounted(() =>
  act(async () => {
    participants.value = await api<Participant[]>('participants');
    const quotas = await Promise.all(
      participants.value.map((p) => api<Quota[]>(`quotas?participant_id=${p.id}`)),
    );
    quotaNames.value = Object.fromEntries(quotas.flat().map((q) => [q.id, q.name]));
    await load();
  }),
);
</script>
<template>
  <section>
    <div class="section-heading"><h2>账单统计</h2></div>
    <ReportFilters :participants="participants" :mode="tab" @change="apply" /><SummaryStats
      :total="total"
    />
    <div class="panel">
      <div class="report-tabs">
        <button
          v-for="t in tabs"
          :key="t.id"
          class="btn btn-sm"
          :class="tab === t.id ? 'btn-primary' : 'btn-ghost'"
          @click="switchTab(t.id)"
        >
          {{ t.label }}
        </button>
      </div>
      <BillsTable
        v-if="displayedTab === 'bills'"
        :bills="bills"
        :names="names"
        :pending="loading"
      /><AuditsTable
        v-else-if="displayedTab === 'audits'"
        :audits="audits"
        :names="names"
        :pending="loading"
      /><PeriodsTable
        v-else-if="displayedTab === 'periods'"
        :periods="history"
        :names="names"
        :pending="loading"
        :quota-names="quotaNames"
      /><template v-else>
        <SelectField
          v-model="group"
          class="stats-group"
          label="分组"
          :options="[
            { value: 'model', label: '按模型' },
            { value: 'participant', label: '按参与者' },
            { value: 'day', label: '按日期' },
          ]"
          @change="act(load)" />
        <ReportTable
          :items="stats"
          :row-key="(s) => `${displayedGroup}:${s.group}`"
          :columns="6"
          empty="暂无消费"
          :pending="loading"
        >
          <template #header>
            <th>分组</th>
            <th>请求数</th>
            <th>输入 Token</th>
            <th>输出 Token</th>
            <th>缓存读取</th>
            <th>消费</th>
          </template>
          <template #row="{ item: s }">
            <td>{{ displayedGroup === 'participant' ? names[s.group] || s.group : s.group }}</td>
            <td><AnimatedValue :value="s.requests" /></td>
            <td><AnimatedValue :value="s.input" /></td>
            <td><AnimatedValue :value="s.output" /></td>
            <td><AnimatedValue :value="s.cache_read" /></td>
            <td class="amount"><AnimatedValue :value="money(s.cost)" /></td>
          </template> </ReportTable
      ></template>
      <TablePagination
        :total="count"
        :page="Math.floor(displayedOffset / PAGE_SIZE) + 1"
        :page-size="PAGE_SIZE"
        :disabled="loading"
        @page="goPage"
      />
    </div>
  </section>
</template>
