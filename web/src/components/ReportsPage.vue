<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import type { Participant, Bill, Period, Audit, Stat, Page, Quota } from '../types';
import { api, act, money, busy } from '../api';
import ReportFilters from './components/ReportFilters.vue';
import SummaryStats from './components/SummaryStats.vue';
import BillsTable from './components/BillsTable.vue';
import PeriodsTable from './components/PeriodsTable.vue';
import AuditsTable from './components/AuditsTable.vue';
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
let filters = new URLSearchParams();
const tabs = [
  { id: 'bills', label: '请求账单' },
  { id: 'stats', label: '消费统计' },
  { id: 'periods', label: '额度周期' },
  { id: 'audits', label: '调整记录' },
];
async function load() {
  loading.value = true;
  try {
    const q = new URLSearchParams(filters);
    q.set('limit', '20');
    q.set('offset', String(offset.value));
    total.value = (await api<Stat[]>(`stats?${filters}`))[0];
    if (tab.value === 'bills') {
      const page = await api<Page<Bill>>(`bills?${q}`);
      bills.value = page.items;
      count.value = page.total;
    } else if (tab.value === 'audits') {
      const page = await api<Page<Audit>>(`audits?${q}`);
      audits.value = page.items;
      count.value = page.total;
    } else if (tab.value === 'periods') {
      history.value = await api<Period[]>(`periods?${q}`);
    } else {
      q.set('group', group.value);
      stats.value = await api<Stat[]>(`stats?${q}`);
    }
  } finally {
    loading.value = false;
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
async function paginate(delta: number) {
  offset.value += delta;
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
      <div v-if="loading" class="empty">读取中</div>
      <template v-else
        ><BillsTable v-if="tab === 'bills'" :bills="bills" :names="names" /><AuditsTable
          v-else-if="tab === 'audits'"
          :audits="audits"
          :names="names"
        /><PeriodsTable
          v-else-if="tab === 'periods'"
          :periods="history"
          :names="names"
          :quota-names="quotaNames"
        /><template v-else
          ><label class="stats-group"
            ><span>分组</span
            ><select v-model="group" class="select" @change="act(load)">
              <option value="model">按模型</option>
              <option value="participant">按参与者</option>
              <option value="day">按日期</option>
            </select></label
          >
          <div v-if="!stats.length" class="empty">暂无消费</div>
          <div v-else class="table-wrap">
            <table class="table">
              <thead>
                <tr>
                  <th>分组</th>
                  <th>请求数</th>
                  <th>输入 Token</th>
                  <th>输出 Token</th>
                  <th>缓存读取</th>
                  <th>消费</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="s in stats" :key="s.group">
                  <td>{{ group === 'participant' ? names[s.group] || s.group : s.group }}</td>
                  <td>{{ s.requests }}</td>
                  <td>{{ s.input }}</td>
                  <td>{{ s.output }}</td>
                  <td>{{ s.cache_read }}</td>
                  <td class="amount">{{ money(s.cost) }}</td>
                </tr>
              </tbody>
            </table>
          </div></template
        ></template
      >
      <div v-if="['bills', 'audits'].includes(tab)" class="pagination">
        <span>{{ count }} 条</span
        ><button class="btn btn-sm" :disabled="offset === 0 || busy" @click="paginate(-20)">
          上一页</button
        ><span>{{ Math.floor(offset / 20) + 1 }}</span
        ><button class="btn btn-sm" :disabled="offset + 20 >= count || busy" @click="paginate(20)">
          下一页
        </button>
      </div>
    </div>
  </section>
</template>
