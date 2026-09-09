<script setup lang="ts">
import type { Period } from '../../types';
import { date, money } from '../../api';
defineProps<{
  periods: Period[];
  names: Record<string, string>;
  quotaNames: Record<string, string>;
}>();
</script>
<template>
  <div v-if="!periods.length" class="empty">暂无额度周期</div>
  <div v-else class="table-wrap table-list table-list-inset">
    <table class="table">
      <thead>
        <tr>
          <th>参与者 / 额度</th>
          <th>开始</th>
          <th>结束</th>
          <th>关闭时间</th>
          <th>限额</th>
          <th>已用</th>
          <th>剩余</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="p in periods" :key="p.id">
          <td>
            {{ names[p.participant_id] || p.participant_id }} /
            {{ quotaNames[p.quota_id] || p.quota_id }}
          </td>
          <td>{{ date(p.starts_at) }}</td>
          <td>{{ date(p.ends_at) }}</td>
          <td>{{ p.closed_at ? date(p.closed_at) : '当前周期' }}</td>
          <td class="amount">{{ money(p.limit) }}</td>
          <td class="amount">{{ money(p.used) }}</td>
          <td class="amount">{{ money(p.remaining) }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
