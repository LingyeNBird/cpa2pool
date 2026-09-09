<script setup lang="ts">
import { ref } from 'vue';
import type { Bill } from '../../types';
import { money, date } from '../../api';
import BillDetail from './BillDetail.vue';
defineProps<{ bills: Bill[]; names: Record<string, string> }>();
const selected = ref<Bill | null>(null);
</script>
<template>
  <div v-if="!bills.length" class="empty">暂无请求账单</div>
  <div v-else class="table-wrap table-list table-list-inset">
    <table class="table">
      <thead>
        <tr>
          <th>时间</th>
          <th>参与者</th>
          <th>模型</th>
          <th>输入 / 输出</th>
          <th>缓存读取</th>
          <th>扣费</th>
          <th>明细</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="b in bills" :key="b.id">
          <td>{{ date(b.time) }}</td>
          <td>{{ names[b.participant_id] || b.participant_id }}</td>
          <td>{{ b.model }}</td>
          <td>{{ b.usage.input }} / {{ b.usage.output }}</td>
          <td>{{ b.usage.cache_read }}</td>
          <td class="amount">{{ money(b.charge.final) }}</td>
          <td><button class="btn btn-sm" @click="selected = b">查看</button></td>
        </tr>
      </tbody>
    </table>
  </div>
  <BillDetail v-if="selected" :bill="selected" @close="selected = null" />
</template>
