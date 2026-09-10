<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import FieldLabel from '../FieldLabel.vue';
import './ModelPicker.css';
const model = defineModel<string[]>({ required: true });
const props = defineProps<{
  label: string;
  tip?: string;
  options: readonly string[];
  disabled?: boolean;
}>();
const trigger = ref<HTMLButtonElement>();
const popup = ref<HTMLElement>();
const open = ref(false);
const summary = computed(() => {
  if (!props.options.length) return '暂无模型';
  return model.value.length === props.options.length
    ? '全部模型'
    : `已选 ${model.value.length} / ${props.options.length}`;
});
function positionPopup() {
  if (!open.value || !trigger.value || !popup.value) return;
  const anchor = trigger.value.getBoundingClientRect();
  const menu = popup.value;
  const gap = 7;
  const left = Math.min(window.innerWidth - menu.offsetWidth - 12, Math.max(12, anchor.left));
  const below = anchor.bottom + gap;
  const top =
    below + menu.offsetHeight <= window.innerHeight - 12
      ? below
      : anchor.top - menu.offsetHeight - gap;
  menu.style.left = `${left}px`;
  menu.style.top = `${Math.max(12, top)}px`;
}
function togglePopup() {
  if (!popup.value || props.disabled || !props.options.length) return;
  popup.value.togglePopover();
}
function updateOpen(event: Event) {
  open.value = (event as ToggleEvent).newState === 'open';
  if (open.value) requestAnimationFrame(positionPopup);
}
function toggleOption(option: string) {
  if (model.value.includes(option)) {
    if (model.value.length === 1) return;
    model.value = model.value.filter((value) => value !== option);
  } else {
    model.value = [...model.value, option];
  }
}
onMounted(() => {
  window.addEventListener('resize', positionPopup);
  document.addEventListener('scroll', positionPopup, true);
});
onBeforeUnmount(() => {
  window.removeEventListener('resize', positionPopup);
  document.removeEventListener('scroll', positionPopup, true);
});
</script>
<template>
  <div class="field model-picker-field">
    <label
      ><FieldLabel v-if="tip" :text="label" :tip="tip" /><span v-else>{{ label }}</span></label
    >
    <button
      ref="trigger"
      type="button"
      class="btn model-picker-trigger"
      :disabled="disabled || !options.length"
      :aria-expanded="open"
      aria-haspopup="dialog"
      @click="togglePopup"
    >
      {{ summary }}
    </button>
    <section
      ref="popup"
      class="model-picker-popup"
      popover="auto"
      aria-label="选择允许模型"
      @beforetoggle="updateOpen"
    >
      <header>
        <strong>允许模型</strong><span>{{ summary }}</span>
      </header>
      <div class="model-picker-options">
        <button
          v-for="option in options"
          :key="option"
          type="button"
          class="model-picker-option"
          :class="{ 'is-selected': model.includes(option) }"
          :aria-pressed="model.includes(option)"
          :title="model.includes(option) && model.length === 1 ? '至少保留一个模型' : undefined"
          @click="toggleOption(option)"
        >
          <span>{{ option }}</span>
        </button>
      </div>
    </section>
  </div>
</template>
