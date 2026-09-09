<script setup lang="ts" generic="T extends string">
import { computed, onMounted, ref, shallowRef, useId } from 'vue';
import {
  SelectRoot,
  SelectTrigger,
  SelectValue,
  SelectPortal,
  SelectContent,
  SelectViewport,
  SelectItem,
  SelectItemText,
  SelectItemIndicator,
} from 'reka-ui';
import { PhCaretDown, PhCheck } from '@phosphor-icons/vue';
import FieldLabel from './FieldLabel.vue';
import './SelectField.css';
const model = defineModel<T>({ required: true });
const props = defineProps<{
  label: string;
  tip?: string;
  options: readonly { value: T; label: string }[];
  disabled?: boolean;
}>();
const emit = defineEmits<{ change: [] }>();
const id = useId();
const root = ref<HTMLElement>();
const portal = shallowRef<HTMLElement>();
const open = ref(false);
// Reka reserves an empty value for its placeholder; indices preserve the API's empty "全部" option.
const selected = computed(() => {
  const index = props.options.findIndex((option) => option.value === model.value);
  return index < 0 ? undefined : String(index);
});
onMounted(() => {
  portal.value = root.value?.closest('dialog') || document.body;
});
function select(value: unknown) {
  if (typeof value !== 'string') return;
  const option = props.options[Number(value)];
  if (option && option.value !== model.value) {
    model.value = option.value;
    emit('change');
  }
}
function closeOnEscape(event: KeyboardEvent) {
  event.preventDefault();
  open.value = false;
}
</script>
<template>
  <div ref="root" class="field select-field">
    <label :for="id"
      ><FieldLabel v-if="tip" :text="label" :tip="tip" /><span v-else>{{ label }}</span></label
    >
    <SelectRoot
      v-model:open="open"
      :model-value="selected"
      :disabled="disabled"
      @update:model-value="select"
    >
      <SelectTrigger :id="id" class="select select-trigger" :aria-label="label">
        <SelectValue placeholder="请选择" /><PhCaretDown :size="17" aria-hidden="true" />
      </SelectTrigger>
      <SelectPortal :to="portal" :disabled="!portal">
        <SelectContent
          class="select-menu"
          position="popper"
          :side-offset="6"
          :collision-padding="12"
          @escape-key-down="closeOnEscape"
        >
          <SelectViewport class="select-options">
            <SelectItem
              v-for="(option, index) in options"
              :key="option.value"
              :value="String(index)"
              class="select-option"
            >
              <SelectItemText>{{ option.label }}</SelectItemText>
              <SelectItemIndicator><PhCheck :size="17" weight="bold" /></SelectItemIndicator>
            </SelectItem>
          </SelectViewport>
        </SelectContent>
      </SelectPortal>
    </SelectRoot>
  </div>
</template>
