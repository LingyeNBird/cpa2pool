<script setup lang="ts">
import ReportTable from './ReportTable.vue';
import { ref } from 'vue';
import type { Audit } from '../../types';
import { date } from '../../api';
import AuditDetail from './AuditDetail.vue';
defineProps<{ audits: Audit[]; names: Record<string, string>; pending?: boolean }>();
const selected = ref<Audit | null>(null);
const actions: Record<string, string> = {
  'participant.save': '保存参与者',
  'participant.delete': '删除参与者',
  'price.save': '保存价格',
  'price.delete': '删除价格',
  'quota.create': '创建额度',
  'quota.save': '修改额度',
  'quota.set_limit': '设置限额',
  'quota.add': '增减额度',
  'quota.set_remaining': '设置剩余',
  'quota.expiry': '修改有效期',
  'quota.reset': '重置已用',
  'quota.new_period': '提前新周期',
  'period.end': '周期结束',
};
</script>
<template>
  <ReportTable
    :items="audits"
    :row-key="(a) => a.id"
    :columns="5"
    empty="暂无调整记录"
    :pending="pending"
  >
    <template #header>
      <th>时间</th>
      <th>参与者</th>
      <th>操作</th>
      <th>备注</th>
      <th>明细</th>
    </template>
    <template #row="{ item: a }">
      <td>{{ date(a.time) }}</td>
      <td>{{ names[a.participant_id] || a.participant_id || '全局' }}</td>
      <td>{{ actions[a.action] || a.action }}</td>
      <td>{{ a.note || '无' }}</td>
      <td><button class="btn btn-sm" @click="selected = a">查看</button></td>
    </template>
  </ReportTable>
  <AuditDetail v-if="selected" :audit="selected" @close="selected = null" />
</template>
