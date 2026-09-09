<script setup lang="ts">
import { ref } from 'vue';
import type { Audit } from '../../types';
import { date } from '../../api';
import AuditDetail from './AuditDetail.vue';
defineProps<{ audits: Audit[]; names: Record<string, string> }>();
const selected = ref<Audit | null>(null);
const actions: Record<string, string> = { 'participant.save': '保存参与者', 'participant.delete': '删除参与者', 'price.save': '保存价格', 'price.delete': '删除价格', 'quota.create': '创建额度', 'quota.save': '修改额度', 'quota.set_limit': '设置限额', 'quota.add': '增减额度', 'quota.set_remaining': '设置剩余', 'quota.expiry': '修改有效期', 'quota.reset': '重置已用', 'quota.new_period': '提前新周期', 'period.end': '周期结束' };
</script>
<template><div v-if="!audits.length" class="empty">暂无调整记录</div><div v-else class="table-wrap"><table class="table"><thead><tr><th>时间</th><th>参与者</th><th>操作</th><th>备注</th><th>明细</th></tr></thead><tbody><tr v-for="a in audits" :key="a.id"><td>{{ date(a.time) }}</td><td>{{ names[a.participant_id] || a.participant_id || '全局' }}</td><td>{{ actions[a.action] || a.action }}</td><td>{{ a.note || '无' }}</td><td><button class="btn btn-sm" @click="selected = a">查看</button></td></tr></tbody></table></div><AuditDetail v-if="selected" :audit="selected" @close="selected = null" /></template>
