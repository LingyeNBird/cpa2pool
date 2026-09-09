<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import NumberField from '../NumberField.vue';
import './CalendarTimeEditor.css';
const props = defineProps<{ hours: number; minutes: number; date: Date | null }>();
const emit = defineEmits<{ confirm: [value: { hours: number; minutes: number }]; cancel: [] }>();
const hours = ref(props.hours);
const minutes = ref(props.minutes);
const root = ref<HTMLElement>();
onMounted(() => root.value?.querySelector('input')?.focus());
const valid = computed(
  () =>
    Number.isInteger(hours.value) &&
    hours.value >= 0 &&
    hours.value <= 23 &&
    Number.isInteger(minutes.value) &&
    minutes.value >= 0 &&
    minutes.value <= 59,
);
function confirm() {
  if (valid.value) emit('confirm', { hours: hours.value, minutes: minutes.value });
}
function keydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    event.preventDefault();
    emit('cancel');
  } else if (event.key === 'Enter' && event.target instanceof HTMLInputElement) {
    event.preventDefault();
    confirm();
  }
}
</script>
<template>
  <section
    ref="root"
    class="calendar-time-editor"
    aria-label="编辑时间"
    @pointerdown.stop
    @mousedown.stop
    @keydown.stop="keydown"
  >
    <header>
      <h3>选择时间</h3>
      <span>{{ date ? date.toLocaleDateString('zh-CN') : '今天' }}</span>
    </header>
    <div class="time-unit">
      <NumberField v-model="hours" label="小时" min="0" max="23" step="1" required />
      <input
        v-model.number="hours"
        class="time-slider"
        type="range"
        min="0"
        max="23"
        step="1"
        aria-label="小时滑条"
      />
    </div>
    <div class="time-unit">
      <NumberField v-model="minutes" label="分钟" min="0" max="59" step="1" required />
      <input
        v-model.number="minutes"
        class="time-slider"
        type="range"
        min="0"
        max="59"
        step="1"
        aria-label="分钟滑条"
      />
    </div>
    <footer>
      <button type="button" class="btn btn-sm" @click="emit('cancel')">取消时间</button>
      <button type="button" class="btn btn-sm btn-primary" :disabled="!valid" @click="confirm">
        确定时间
      </button>
    </footer>
  </section>
</template>
