<script setup lang="ts">
import { reactive, ref } from 'vue';
import type { Price } from '../../types';
import { api, busy } from '../../api';
import DialogFrame from './DialogFrame.vue';
import FieldLabel from './FieldLabel.vue';
import NumberField from './NumberField.vue';
import BinaryChoiceField from './components/BinaryChoiceField.vue';
const props = defineProps<{ price: Price | null }>();
const emit = defineEmits<{ close: []; saved: [] }>();
const draft = reactive<Price>({
  ...(props.price || {
    model: '',
    input: '0',
    output: '0',
    cache_read: '0',
    cache_write: '0',
    priority_enabled: false,
    priority_multiplier: '2',
    long_enabled: false,
    long_threshold: 200000,
    long_input_multiplier: '2',
    long_output_multiplier: '2',
    model_enabled: false,
    model_multiplier: '1',
    combination: 'multiply',
    updated_at: '',
  }),
});
const error = ref('');
async function save() {
  busy.value = true;
  error.value = '';
  try {
    await api('prices', props.price ? 'PUT' : 'POST', draft);
    emit('saved');
  } catch (e) {
    error.value = String(e);
  } finally {
    busy.value = false;
  }
}
const rates = [
  { key: 'input', label: '输入' },
  { key: 'output', label: '输出' },
  { key: 'cache_read', label: '缓存读取' },
  { key: 'cache_write', label: '缓存写入' },
] as const;
</script>
<template>
  <DialogFrame :title="price ? '编辑计费' : '添加模型'" @close="emit('close')"
    ><form @submit.prevent="save">
      <div v-if="error" class="alert alert-error error-bar">{{ error }}</div>
      <div class="form-grid">
        <label class="field full-width"
          ><FieldLabel
            text="模型"
            tip="填写上游实际模型名称。修改规则仅影响新请求，历史消费保持原计价快照。" /><input
            v-model="draft.model"
            class="input"
            :readonly="!!price"
            required
        /></label>
        <NumberField
          v-for="rate in rates"
          :key="rate.key"
          v-model="draft[rate.key]"
          :label="`${rate.label}单价`"
          tip="USD / 百万 Token。缓存读取与普通输入分开计算，推理 Token 不重复加价。"
          min="0"
          required
        />
        <div class="full-width"><hr class="divider-line" /></div>
        <label class="check-field"
          ><input
            v-model="draft.priority_enabled"
            class="toggle toggle-primary"
            type="checkbox"
          />FAST / priority</label
        >
        <NumberField
          v-model="draft.priority_multiplier"
          label="倍率"
          min="0.000001"
          :disabled="!draft.priority_enabled"
        />
        <label class="check-field"
          ><input
            v-model="draft.long_enabled"
            class="toggle toggle-primary"
            type="checkbox"
          />长上下文</label
        >
        <NumberField
          v-model.number="draft.long_threshold"
          label="Token 阈值"
          tip="总输入（含缓存）超过阈值时，整笔请求采用长上下文倍率。"
          min="0"
          step="1"
          :disabled="!draft.long_enabled"
        />
        <NumberField
          v-model="draft.long_input_multiplier"
          label="长上下文输入倍率"
          min="0.000001"
          :disabled="!draft.long_enabled"
        />
        <NumberField
          v-model="draft.long_output_multiplier"
          label="长上下文输出倍率"
          min="0.000001"
          :disabled="!draft.long_enabled"
        />
        <label class="check-field"
          ><input
            v-model="draft.model_enabled"
            class="toggle toggle-primary"
            type="checkbox"
          />模型额外倍率</label
        >
        <NumberField
          v-model="draft.model_multiplier"
          label="倍率"
          min="0.000001"
          :disabled="!draft.model_enabled"
        />
        <BinaryChoiceField
          v-model="draft.combination"
          class="full-width"
          label="组合方式"
          tip="相乘：所有命中倍率相乘。最大值：输入与输出分别取命中倍率中的最大值。"
          :options="[
            { value: 'multiply', label: '相乘' },
            { value: 'max', label: '取最大值' },
          ]"
        />
      </div>
      <div class="form-actions">
        <button type="button" class="btn" @click="emit('close')">取消</button
        ><button class="btn btn-primary" :disabled="busy">保存</button>
      </div>
    </form></DialogFrame
  >
</template>
