<script setup lang="ts">
import ReportTable from './ReportTable.vue';
import AnimatedValue from './AnimatedValue.vue';
import type { Period } from '../../types';
import { date, money } from '../../api';
defineProps<{
  periods: Period[];
  names: Record<string, string>;
  quotaNames: Record<string, string>;
  pending?: boolean;
}>();
</script>
<template>
  <ReportTable
    :items="periods"
    :row-key="(p) => p.id"
    :columns="7"
    empty="暂无额度周期"
    :pending="pending"
  >
    <template #header>
      <th>参与者 / 额度</th>
      <th>开始</th>
      <th>结束</th>
      <th>关闭时间</th>
      <th>限额</th>
      <th>已用</th>
      <th>剩余</th>
    </template>
    <template #row="{ item: p }">
      <td>
        {{ names[p.participant_id] || p.participant_id }} /
        {{ quotaNames[p.quota_id] || p.quota_id }}
      </td>
      <td>{{ date(p.starts_at) }}</td>
      <td>{{ date(p.ends_at) }}</td>
      <td>{{ p.closed_at ? date(p.closed_at) : '当前周期' }}</td>
      <td class="amount"><AnimatedValue :value="money(p.limit)" /></td>
      <td class="amount"><AnimatedValue :value="money(p.used)" /></td>
      <td class="amount"><AnimatedValue :value="money(p.remaining)" /></td>
    </template>
  </ReportTable>
</template>
