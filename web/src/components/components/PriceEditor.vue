<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import type { Price } from '../../types';
import { api, busy } from '../../api';
import DialogFrame from './DialogFrame.vue';
import FieldLabel from './FieldLabel.vue';
import NumberField from './NumberField.vue';
import BinaryChoiceField from './components/BinaryChoiceField.vue';
import SelectField from './SelectField.vue';
import { loadModelCatalog, type ModelCatalogItem } from '../../modelCatalog';
const props = defineProps<{ price: Price | null; mode?: 'token' | 'image' }>();
const emit = defineEmits<{ close: []; saved: [] }>();
const draft = reactive<Price>(
  props.price
    ? {
        ...props.price,
        billing_mode: props.price.billing_mode || 'token',
        image_price_1k: props.price.image_price_1k || '0',
        image_price_2k: props.price.image_price_2k || '0',
        image_price_4k: props.price.image_price_4k || '0',
      }
    : {
        model: '',
        billing_mode: 'token',
        image_price_1k: '0',
        image_price_2k: '0',
        image_price_4k: '0',
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
      },
);
const error = ref('');
const catalog = ref<ModelCatalogItem[]>([]);
const loadingCatalog = ref(!props.price);
const catalogError = ref('');
const modelOptions = computed(() =>
  catalog.value.map((item) => ({
    value: item.id,
    label:
      item.source === 'models.dev'
        ? item.id
        : `${item.id} · ${item.source === 'sub2api' ? 'Sub2API 价格' : 'Sub2API 回退价格'}`,
  })),
);
function applyCatalogPrice() {
  const item = catalog.value.find((candidate) => candidate.id === draft.model);
  if (!item) return;
  Object.assign(draft, item.price);
}
async function loadCatalog() {
  if (props.price) return;
  try {
    catalog.value = await loadModelCatalog();
    if (!catalog.value.length) {
      catalogError.value = 'CPA 当前没有返回可用模型，请手动填写模型名称和价格。';
      return;
    }
    draft.model = catalog.value[0].id;
    applyCatalogPrice();
  } catch (e) {
    catalogError.value = e instanceof Error ? e.message : String(e);
  } finally {
    loadingCatalog.value = false;
  }
}
onMounted(loadCatalog);
async function save() {
  draft.billing_mode = props.mode === 'image' ? 'image' : draft.billing_mode || 'token';
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
  <DialogFrame
    :title="price ? (mode === 'image' ? '编辑图片计费' : '编辑 Token 计费') : '添加模型'"
    @close="emit('close')"
    ><form @submit.prevent="save">
      <div v-if="error" class="alert alert-error error-bar">{{ error }}</div>
      <div class="form-grid">
        <label v-if="price || catalogError" class="field full-width"
          ><FieldLabel
            text="模型"
            tip="模型目录读取失败时可手动填写；修改规则仅影响新请求，历史消费保持原计价快照。" /><input
            v-model="draft.model"
            class="input"
            :readonly="!!price"
            required
        /></label>
        <SelectField
          v-else
          v-model="draft.model"
          class="full-width"
          label="模型"
          tip="来自 CPA 当前所有渠道的可用模型；优先使用 models.dev，缺失时采用 Sub2API 价格与回退规则。"
          :options="modelOptions"
          :disabled="loadingCatalog"
          @change="applyCatalogPrice"
        />
        <div v-if="catalogError && !price" class="alert full-width">{{ catalogError }}</div>
        <template v-if="mode === 'image'">
          <NumberField
            v-model="draft.image_price_1k"
            label="1K 每张价格"
            tip="USD / 张。默认采用 Sub2API 图片计费费率。"
            min="0"
            required
          />
          <NumberField
            v-model="draft.image_price_2k"
            label="2K 每张价格"
            tip="USD / 张。"
            min="0"
            required
          />
          <NumberField
            v-model="draft.image_price_4k"
            label="4K 每张价格"
            tip="USD / 张。"
            min="0"
            required
          />
        </template>
        <template v-else>
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
        </template>
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
          v-if="mode !== 'image'"
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
        ><button class="btn btn-primary" :disabled="busy || loadingCatalog">保存</button>
      </div>
    </form></DialogFrame
  >
</template>
