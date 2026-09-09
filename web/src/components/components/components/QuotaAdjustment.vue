<script setup lang="ts">
import { ref } from 'vue';
import type { QuotaView } from '../../../types';
import { api, busy, money } from '../../../api';
import DialogFrame from '../DialogFrame.vue';
import FieldLabel from '../FieldLabel.vue';
const props = defineProps<{ quota: QuotaView }>();
const emit = defineEmits<{ close: []; saved: [] }>();
const action = ref('add'); const amount = ref('10'); const restart = ref(false); const note = ref(''); const error = ref('');
const tips: Record<string, string> = { add: '正数增加额度，负数减少额度；已用量不变。', set_limit: '直接修改限额总额，剩余 = 新限额 − 已用。', set_remaining: '直接设置当前剩余，限额 = 已用 + 新剩余。', reset: '保留历史记录并清零当前已用；可选择重新开始周期计时。', new_period: '保留历史记录，清零已用，并从现在开始计时。' };
async function save() { busy.value = true; error.value = ''; try { await api('adjustments', 'POST', { quota_id: props.quota.id, action: action.value, amount: amount.value, restart: restart.value, note: note.value }); emit('saved'); } catch (e) { error.value = String(e); } finally { busy.value = false; } }
</script>
<template><DialogFrame title="调整额度" @close="emit('close')"><form @submit.prevent="save"><div v-if="error" class="alert alert-error error-bar">{{ error }}</div><div class="section-heading"><h3>{{ quota.name }}</h3><span class="amount">{{ money(quota.current.remaining) }}</span></div><div class="form-grid">
<label class="field full-width"><FieldLabel text="操作" :tip="tips[action]" /><select v-model="action" class="select"><option value="add">增减额度</option><option value="set_limit">设置限额总额</option><option value="set_remaining">设置当前剩余</option><option value="reset">重置已用量</option><option value="new_period">提前开始新周期</option></select></label>
<label v-if="!['reset', 'new_period'].includes(action)" class="field full-width"><span>金额 / USD</span><input v-model="amount" class="input" type="number" :min="action === 'add' ? undefined : 0" step="any" required /></label>
<label v-if="action === 'reset'" class="check-field full-width"><input v-model="restart" class="checkbox checkbox-primary" type="checkbox" />同时重新开始周期计时</label>
<label class="field full-width"><span>备注</span><input v-model="note" class="input" /></label></div><div class="form-actions"><button type="button" class="btn" @click="emit('close')">取消</button><button class="btn btn-primary" :disabled="busy">确认调整</button></div></form></DialogFrame></template>
