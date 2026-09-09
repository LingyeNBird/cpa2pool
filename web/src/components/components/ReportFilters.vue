<script setup lang="ts">
import { ref, watch } from 'vue';
import type { Participant } from '../../types';
import { iso } from '../../api';
import './ReportFilters.css';
import DateTimeField from './DateTimeField.vue';
const props = defineProps<{ participants: Participant[]; mode: string }>();
const emit = defineEmits<{ change: [params: URLSearchParams] }>();
const pid = ref('');
const model = ref('');
const from = ref('');
const to = ref('');
watch(
  () => props.mode,
  (mode) => {
    if (mode === 'periods' || mode === 'audits') model.value = '';
    if (mode === 'periods') {
      from.value = '';
      to.value = '';
    }
  },
);
function apply() {
  const q = new URLSearchParams();
  if (pid.value) q.set('participant_id', pid.value);
  if (model.value) q.set('model', model.value);
  if (from.value) q.set('from', iso(from.value)!);
  if (to.value) q.set('to', iso(to.value)!);
  emit('change', q);
}
</script>
<template>
  <form class="report-filters" @submit.prevent="apply">
    <label class="field"
      ><span>参与者</span
      ><select v-model="pid" class="select">
        <option value="">全部</option>
        <option v-for="p in participants" :key="p.id" :value="p.id">{{ p.name }}</option>
      </select></label
    ><label v-if="mode === 'bills' || mode === 'stats'" class="field"
      ><span>模型</span><input v-model="model" class="input" placeholder="全部"
    /></label>
    <DateTimeField v-if="mode !== 'periods'" v-model="from" label="开始时间" />
    <DateTimeField v-if="mode !== 'periods'" v-model="to" label="结束时间" />
    <button class="btn btn-primary">查询</button>
  </form>
</template>
