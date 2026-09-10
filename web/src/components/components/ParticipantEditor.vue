<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import type { Participant, Price } from '../../types';
import { api, busy, cpaManagement, localDate, iso } from '../../api';
import DialogFrame from './DialogFrame.vue';
import FieldLabel from './FieldLabel.vue';
import SelectField from './SelectField.vue';
import ModelPicker from './components/ModelPicker.vue';
import './ParticipantEditor.css';
import DateTimeField from './DateTimeField.vue';
const props = defineProps<{ participant: Participant | null }>();
const emit = defineEmits<{ close: []; saved: [] }>();
const p = props.participant;
const draft = reactive({
  ...(p || { id: '', name: '', note: '', enabled: true, paused: false }),
  api_key: '',
});
const availableModels = ref<string[]>([]);
const selectedModels = ref<string[]>([]);
const loadingModels = ref(true);
const efforts = ref(p?.efforts.join(', ') || '');
const expiry = ref(localDate(p?.expires_at));
const error = ref('');
const keys = ref<string[]>([]);
const loadingKeys = ref(true);
const creatingKey = ref(false);
const keyOptions = computed(() =>
  keys.value.map((key) => ({
    value: key,
    label: `${key.slice(0, 3)}••••${key.slice(-6)}`,
  })),
);
const apiKeyCharset = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
function generateAPIKey() {
  const characters: string[] = [];
  const unbiasedLimit = Math.floor(256 / apiKeyCharset.length) * apiKeyCharset.length;
  while (characters.length < 48) {
    const bytes = crypto.getRandomValues(new Uint8Array(56));
    for (const byte of bytes) {
      if (byte < unbiasedLimit) characters.push(apiKeyCharset[byte % apiKeyCharset.length]);
      if (characters.length === 48) break;
    }
  }
  return `sk-${characters.join('')}`;
}
async function loadKeys() {
  loadingKeys.value = true;
  error.value = '';
  try {
    const configured = await cpaManagement<{ 'api-keys': string[] }>('api-keys');
    keys.value = await api<string[]>('participant-key-candidates', 'POST', {
      keys: configured['api-keys'] || [],
      current_id: p?.id || '',
    });
    if (keys.value.length) draft.api_key = keys.value[0];
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e);
  } finally {
    loadingKeys.value = false;
  }
}
async function loadModels() {
  loadingModels.value = true;
  try {
    const prices = await api<Price[]>('prices');
    availableModels.value = prices.map((price) => price.model);
    const configured = p?.models || [];
    selectedModels.value = configured.length
      ? availableModels.value.filter((model) => configured.includes(model))
      : [...availableModels.value];
    if (!selectedModels.value.length && availableModels.value.length) {
      selectedModels.value = [...availableModels.value];
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e);
  } finally {
    loadingModels.value = false;
  }
}
async function createKey() {
  creatingKey.value = true;
  error.value = '';
  try {
    const key = generateAPIKey();
    await cpaManagement('api-keys', 'PATCH', {
      old: `cpa2pool-missing-${key}`,
      new: key,
    });
    keys.value = [key];
    draft.api_key = key;
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e);
  } finally {
    creatingKey.value = false;
  }
}
onMounted(() => {
  loadKeys();
  loadModels();
});
async function save() {
  busy.value = true;
  error.value = '';
  try {
    await api('participants', p ? 'PUT' : 'POST', {
      ...draft,
      models:
        selectedModels.value.length === availableModels.value.length ? [] : selectedModels.value,
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
        <SelectField
          v-if="keys.length"
          v-model="draft.api_key"
          label="API Key"
          tip="仅显示 CPA 中尚未关联其他参与者的 API Key。"
          :options="keyOptions"
        />
        <div v-else class="field">
          <FieldLabel text="API Key" tip="CPA 中没有尚未关联参与者的 API Key。" />
          <button
            type="button"
            class="btn participant-flat-action"
            :disabled="loadingKeys || creatingKey"
            @click="createKey"
          >
            {{ loadingKeys ? '读取中' : creatingKey ? '创建中' : '创建 API Key' }}
          </button>
        </div>
        <label class="field full-width"
          ><span>备注</span><textarea v-model="draft.note" class="textarea" rows="2" />
        </label>
        <ModelPicker
          v-model="selectedModels"
          label="允许模型"
          tip="至少保留一个模型；选择全部时允许所有已定价模型。"
          :options="availableModels"
          :disabled="loadingModels"
        />
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
        ><button
          class="btn btn-primary"
          :disabled="busy || loadingKeys || loadingModels || (!p && !draft.api_key)"
        >
          保存
        </button>
      </div>
    </form></DialogFrame
  >
</template>
