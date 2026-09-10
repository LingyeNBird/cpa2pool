<script setup lang="ts" generic="T extends string">
import FieldLabel from '../FieldLabel.vue';
import './BinaryChoiceField.css';
const model = defineModel<T>({ required: true });
defineProps<{
  label: string;
  tip?: string;
  options: readonly [{ value: T; label: string }, { value: T; label: string }];
}>();
</script>
<template>
  <div class="field binary-choice-field">
    <label
      ><FieldLabel v-if="tip" :text="label" :tip="tip" /><span v-else>{{ label }}</span></label
    >
    <div class="binary-choice" role="group" :aria-label="label">
      <button
        v-for="option in options"
        :key="option.value"
        type="button"
        class="binary-choice-key"
        :class="{ 'is-selected': model === option.value }"
        :aria-pressed="model === option.value"
        @click="model = option.value"
      >
        <span>{{ option.label }}</span>
      </button>
    </div>
  </div>
</template>
