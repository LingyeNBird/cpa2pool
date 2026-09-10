<script setup lang="ts">
import { nextTick, ref, useId } from 'vue';
import { PhQuestion } from '@phosphor-icons/vue';
import './TooltipIcon.css';
defineProps<{ tip: string; label?: string; size?: number }>();
const id = useId();
const trigger = ref<HTMLButtonElement>();
const popup = ref<HTMLElement>();
async function show() {
  popup.value?.showPopover();
  await nextTick();
  const anchor = trigger.value?.getBoundingClientRect();
  const tooltip = popup.value;
  if (!anchor || !tooltip) return;
  const gap = 7;
  const left = Math.min(
    window.innerWidth - tooltip.offsetWidth - 8,
    Math.max(8, anchor.left + anchor.width / 2 - tooltip.offsetWidth / 2),
  );
  tooltip.style.left = `${left}px`;
  tooltip.style.top = `${anchor.bottom + gap}px`;
}
function hide() {
  popup.value?.hidePopover();
}
</script>
<template>
  <span class="tooltip-icon">
    <button
      ref="trigger"
      type="button"
      class="tooltip-trigger"
      :aria-label="label || tip"
      :aria-describedby="id"
      @mouseenter="show"
      @mouseleave="hide"
      @focus="show"
      @blur="hide"
    >
      <PhQuestion :size="size || 16" />
    </button>
    <span :id="id" ref="popup" class="tooltip-popup" role="tooltip" popover="manual">{{
      tip
    }}</span>
  </span>
</template>
