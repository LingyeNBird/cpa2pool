<script setup lang="ts">
import { reactive, ref } from 'vue';
import type { Participant } from '../../types';
import { api, busy, localDate, iso } from '../../api';
import DialogFrame from './DialogFrame.vue';
import FieldLabel from './FieldLabel.vue';
import DateTimeField from './DateTimeField.vue';
const props = defineProps<{ participant: Participant | null }>();
const emit = defineEmits<{ close: []; saved: [] }>();
const p = props.participant;
const draft = reactive({
  ...(p || { id: '', name: '', note: '', enabled: true, paused: false }),
  api_key: '',
});
const models = ref(p?.models.join(', ') || '');
const efforts = ref(p?.efforts.join(', ') || '');
const expiry = ref(localDate(p?.expires_at));
const error = ref('');
async function save() {
  busy.value = true;
  error.value = '';
  try {
    await api('participants', p ? 'PUT' : 'POST', {
      ...draft,
      models: models.value
        .split(',')
        .map((v) => v.trim())
        .filter(Boolean),
      efforts: efforts.value
        .split(',')
        .map((v) => v.trim())
        .filter(Boolean),
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
  <DialogFrame :title="p ? '编辑参与者' : '添加参与者'" @close="emit('close')"
    ><form @submit.prevent="save">
      <div v-if="error" class="alert alert-error error-bar">{{ error }}</div>
      <div class="form-grid">
        <label class="field"
          ><span>名称</span><input v-model="draft.name" class="input" required
        /></label>
        <label class="field"
          ><FieldLabel
            text="API Key"
            tip="关联已在 CPA 配置的 API Key。更换密钥不改变参与者标识；编辑时留空则保留。" /><input
            v-model="draft.api_key"
            class="input"
            type="password"
            autocomplete="new-password"
            :required="!p"
            :placeholder="p?.key_preview"
        /></label>
        <label class="field full-width"
          ><span>备注</span><textarea v-model="draft.note" class="textarea" rows="2" />
        </label>
        <label class="field"
          ><FieldLabel
            text="允许模型"
            tip="逗号分隔客户端模型名称，留空允许全部已定价模型。" /><input
            v-model="models"
            class="input"
        /></label>
        <label class="field"
          ><FieldLabel
            text="推理强度"
            tip="逗号分隔，例如 default, low, medium, high, xhigh。default 表示请求未指定强度，留空不限制。" /><input
            v-model="efforts"
            class="input"
        /></label>
        <DateTimeField
          v-model="expiry"
          class="full-width"
          label="有效期"
          tip="留空表示参与者长期有效，额度仍按各自有效期限制。"
        />
        <label class="check-field"
          ><input
            v-model="draft.enabled"
            class="toggle toggle-primary"
            type="checkbox"
          />启用</label
        ><label class="check-field"
          ><input v-model="draft.paused" class="toggle toggle-primary" type="checkbox" />暂停</label
        >
      </div>
      <div class="form-actions">
        <button type="button" class="btn" @click="emit('close')">取消</button
        ><button class="btn btn-primary" :disabled="busy">保存</button>
      </div>
    </form></DialogFrame
  >
</template>
