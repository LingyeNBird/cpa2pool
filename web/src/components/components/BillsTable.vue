<script setup lang="ts">
import ReportTable from './ReportTable.vue';
import AnimatedValue from './AnimatedValue.vue';
import { ref } from 'vue';
import type { Bill } from '../../types';
import { money, date } from '../../api';
import BillDetail from './BillDetail.vue';
defineProps<{ bills: Bill[]; names: Record<string, string>; pending?: boolean }>();
const selected = ref<Bill | null>(null);
</script>
<template>
  <ReportTable
    :items="bills"
    :row-key="(b) => b.id"
    :columns="7"
    empty="暂无请求账单"
    :pending="pending"
  >
    <template #header>
      <th>时间</th>
      <th>参与者</th>
      <th>模型</th>
      <th>输入 / 输出</th>
      <th>缓存读取</th>
      <th>扣费</th>
      <th>明细</th>
    </template>
    <template #row="{ item: b }">
      <td>{{ date(b.time) }}</td>
      <td>{{ names[b.participant_id] || b.participant_id }}</td>
      <td>{{ b.model }}</td>
      <td><AnimatedValue :value="b.usage.input" /> / <AnimatedValue :value="b.usage.output" /></td>
      <td><AnimatedValue :value="b.usage.cache_read" /></td>
      <td class="amount"><AnimatedValue :value="money(b.charge.final)" /></td>
      <td><button class="btn btn-sm" @click="selected = b">查看</button></td>
    </template>
  </ReportTable>
  <BillDetail v-if="selected" :bill="selected" @close="selected = null" />
</template>
