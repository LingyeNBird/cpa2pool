<script setup lang="ts">
import { ref } from 'vue';
import type { Participant } from '../../types';
import { iso } from '../../api';
import './ReportFilters.css';
defineProps<{ participants: Participant[] }>();
const emit = defineEmits<{ change: [params: URLSearchParams] }>();
const pid = ref(''); const model = ref(''); const from = ref(''); const to = ref('');
function apply() { const q = new URLSearchParams(); if (pid.value) q.set('participant_id', pid.value); if (model.value) q.set('model', model.value); if (from.value) q.set('from', iso(from.value)!); if (to.value) q.set('to', iso(to.value)!); emit('change', q); }
</script>
<template><form class="report-filters" @submit.prevent="apply"><label class="field"><span>参与者</span><select v-model="pid" class="select"><option value="">全部</option><option v-for="p in participants" :key="p.id" :value="p.id">{{ p.name }}</option></select></label><label class="field"><span>模型</span><input v-model="model" class="input" placeholder="全部" /></label><label class="field"><span>开始时间</span><input v-model="from" class="input" type="datetime-local" /></label><label class="field"><span>结束时间</span><input v-model="to" class="input" type="datetime-local" /></label><button class="btn btn-primary">查询</button></form></template>
