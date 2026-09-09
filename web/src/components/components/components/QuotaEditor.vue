<script setup lang="ts">
import { reactive, ref } from 'vue';
import type { Quota } from '../../../types';
import { api, busy, iso, localDate } from '../../../api';
import DialogFrame from '../DialogFrame.vue';
import FieldLabel from '../FieldLabel.vue';
import DateTimeField from '../DateTimeField.vue';
const props = defineProps<{ participantId: string; quota: Quota | null }>();
const emit = defineEmits<{ close: []; saved: [] }>();
const q = props.quota;
const draft = reactive({
  ...(q || {
    id: '',
    participant_id: props.participantId,
    name: '',
    limit: '10',
    period: 'week',
    enabled: true,
  }),
});
const start = ref(localDate(q?.starts_at || new Date().toISOString()));
const expiry = ref(localDate(q?.expires_at));
const error = ref('');
async function save() {
  busy.value = true;
  error.value = '';
  try {
    await api('quotas', q ? 'PUT' : 'POST', {
      ...draft,
      starts_at: q?.starts_at || iso(start.value),
      expires_at: iso(expiry.value),
    });
    emit('saved');
  } catch (e) {
    error.value = String(e);
  } finally {
    busy.value = false;
  }
}
</script>
<template>
  <DialogFrame :title="q ? '编辑额度' : '添加额度'" @close="emit('close')"
    ><form @submit.prevent="save">
      <div v-if="error" class="alert alert-error error-bar">{{ error }}</div>
      <div class="form-grid">
        <label class="field"
          ><span>名称</span><input v-model="draft.name" class="input" required /></label
        ><label class="field"
          ><FieldLabel text="限额总额 / USD" tip="修改限额不清空已用量，也不改变当前周期。" /><input
            v-model="draft.limit"
            class="input"
            type="number"
            min="0"
            step="any"
            required
        /></label>
        <label class="field"
          ><span>周期</span
          ><select v-model="draft.period" class="select" :disabled="!!q">
            <option value="none">长期余额</option>
            <option value="day">每日</option>
            <option value="week">每周</option>
            <option value="month">每月</option>
          </select></label
        >
        <DateTimeField v-model="start" label="开始时间" :disabled="!!q" required />
        <DateTimeField
          v-model="expiry"
          class="full-width"
          label="有效期"
          tip="留空表示长期有效。多个生效额度同时约束请求，任一用尽即停止使用。"
        />
        <label class="check-field"
          ><input
            v-model="draft.enabled"
            class="toggle toggle-primary"
            type="checkbox"
          />启用额度</label
        >
      </div>
      <div class="form-actions">
        <button type="button" class="btn" @click="emit('close')">取消</button
        ><button class="btn btn-primary" :disabled="busy">保存</button>
      </div>
    </form></DialogFrame
  >
</template>
