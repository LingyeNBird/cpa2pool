<script setup lang="ts" generic="T extends string | number">
import { computed, useId } from 'vue';
import Big from 'big.js';
import { PhMinus, PhPlus } from '@phosphor-icons/vue';
import FieldLabel from './FieldLabel.vue';
import './NumberField.css';
const model = defineModel<T>({ required: true });
const props = withDefaults(
  defineProps<{
    label: string;
    tip?: string;
    min?: string | number;
    max?: string | number;
    step?: string | number;
    disabled?: boolean;
    required?: boolean;
  }>(),
  { step: 'any' },
);
const id = useId();
const numericModel = typeof model.value === 'number';
const current = computed(() => new Big(model.value || 0));
const decreaseDisabled = computed(
  () => props.disabled || (props.min !== undefined && current.value.lte(props.min)),
);
const increaseDisabled = computed(
  () => props.disabled || (props.max !== undefined && current.value.gte(props.max)),
);
function update(value: string) {
  model.value = (numericModel && value !== '' ? Number(value) : value) as T;
}
function adjust(direction: number) {
  if (direction < 0 ? decreaseDisabled.value : increaseDisabled.value) return;
  const increment = props.step === 'any' ? 1 : props.step;
  let value = current.value.plus(new Big(increment).times(direction));
  if (props.min !== undefined && value.lt(props.min)) value = new Big(props.min);
  if (props.max !== undefined && value.gt(props.max)) value = new Big(props.max);
  update(value.toFixed());
}
</script>
<template>
  <div class="field number-field">
    <label :for="id"
      ><FieldLabel v-if="tip" :text="label" :tip="tip" /><span v-else>{{ label }}</span></label
    >
    <div class="number-control" :data-disabled="disabled || undefined">
      <button
        type="button"
        class="number-step"
        :aria-label="`${label}减小`"
        :disabled="decreaseDisabled"
        @click="adjust(-1)"
      >
        <PhMinus :size="17" weight="bold" />
      </button>
      <input
        :id="id"
        :value="model"
        class="input number-value"
        type="number"
        :min="min"
        :max="max"
        :step="step"
        :disabled="disabled"
        :required="required"
        @input="update(($event.target as HTMLInputElement).value)"
        @keydown.up.prevent="adjust(1)"
        @keydown.down.prevent="adjust(-1)"
      />
      <button
        type="button"
        class="number-step"
        :aria-label="`${label}增大`"
        :disabled="increaseDisabled"
        @click="adjust(1)"
      >
        <PhPlus :size="17" weight="bold" />
      </button>
    </div>
  </div>
</template>
