<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, shallowRef, useId } from 'vue';
import {
  VueDatePicker,
  type InputParsedDate,
  type InternalModelValue,
  type ModelValue,
  type TimeOverlaySlotProps,
} from '@vuepic/vue-datepicker';
import { format, isValid, parse } from 'date-fns';
import { zhCN } from 'date-fns/locale';
import { PhCalendarBlank, PhClock } from '@phosphor-icons/vue';
import FieldLabel from './FieldLabel.vue';
import CalendarTimeEditor from './components/CalendarTimeEditor.vue';
import '@vuepic/vue-datepicker/dist/main.css';
import './DateTimeField.css';
const model = defineModel<string>({ required: true });
defineProps<{ label: string; tip?: string; disabled?: boolean; required?: boolean }>();
const id = useId();
const root = ref<HTMLElement>();
const picker = ref<InstanceType<typeof VueDatePicker>>();
const portal = shallowRef<HTMLElement>();
const hostDialog = shallowRef<HTMLDialogElement>();
const open = ref(false);
const timeEditorOpen = ref(false);
const selectedTime = ref('');
const selectedDate = shallowRef<Date | null>(null);
const timePreview = ref<HTMLElement>();
const inputFormat = 'yyyy-MM-dd HH:mm';
onMounted(() => {
  hostDialog.value = root.value?.closest('dialog') || undefined;
  portal.value =
    hostDialog.value?.querySelector<HTMLElement>('.dialog-portal-host') || document.body;
  hostDialog.value?.addEventListener('cancel', dismissCalendarFirst, true);
});
onBeforeUnmount(() => hostDialog.value?.removeEventListener('cancel', dismissCalendarFirst, true));
function dismissCalendarFirst(event: Event) {
  if (!open.value) return;
  event.preventDefault();
  event.stopImmediatePropagation();
  picker.value?.closeMenu();
}
function preventDialogCancel(event: KeyboardEvent) {
  if (open.value && event.key === 'Escape') event.preventDefault();
}
function parseInput(value: string) {
  const date = parse(value, inputFormat, new Date());
  return isValid(date) && format(date, inputFormat) === value ? date : null;
}
function validateInput(_event: Event | string, parsedDate: InputParsedDate) {
  const input = root.value?.querySelector('input');
  input?.setCustomValidity(
    input.value && !parsedDate ? '请输入有效日期和时间，格式为 YYYY-MM-DD HH:mm。' : '',
  );
}
function update(value: ModelValue) {
  model.value = typeof value === 'string' ? value : '';
  root.value?.querySelector('input')?.setCustomValidity('');
}
function updateTimePreview(value: InternalModelValue) {
  selectedDate.value = value instanceof Date && isValid(value) ? value : null;
  selectedTime.value = selectedDate.value ? format(selectedDate.value, 'HH:mm') : '';
}
function confirmTime(value: { hours: number; minutes: number }, time: TimeOverlaySlotProps) {
  const date = new Date(selectedDate.value || new Date());
  date.setHours(value.hours, value.minutes, 0, 0);
  picker.value?.updateInternalModelValue(date);
  time.setHours(value.hours);
  time.setMinutes(value.minutes);
  time.setSeconds(0);
  closeTimeEditor();
}
function closeTimeEditor() {
  picker.value?.switchView('calendar');
  nextTick(() => timePreview.value?.closest('button')?.focus());
}
function updateOverlayState({ open: overlayOpen, overlay }: { open: boolean; overlay: string }) {
  if (overlay === 'time') timeEditorOpen.value = overlayOpen;
}
</script>
<template>
  <div ref="root" class="field date-time-field" @keydown.esc.capture="preventDialogCancel">
    <label :for="id"
      ><FieldLabel v-if="tip" :text="label" :tip="tip" /><span v-else>{{ label }}</span></label
    >
    <VueDatePicker
      ref="picker"
      :model-value="model || null"
      model-type="yyyy-MM-dd'T'HH:mm"
      :locale="zhCN"
      :disabled="disabled"
      :input-attrs="{ id, required, autocomplete: 'off', clearable: !required }"
      :formats="{ input: inputFormat, preview: inputFormat }"
      :text-input="{
        format: parseInput,
        openMenu: false,
        enterSubmit: true,
        tabSubmit: true,
        applyOnBlur: true,
      }"
      :time-config="{ is24: true, enableSeconds: false }"
      :action-row="{
        selectBtnLabel: '确定',
        cancelBtnLabel: '取消',
        nowBtnLabel: '此刻',
        showNow: true,
        showPreview: false,
      }"
      :aria-labels="{
        input: label,
        menu: `${label}选择器`,
        calendarIcon: '选择日期',
        clearInput: '清空日期',
        nextMonth: '下个月',
        prevMonth: '上个月',
        nextYear: '下一年',
        prevYear: '上一年',
        openYearsOverlay: '选择年份',
        openMonthsOverlay: '选择月份',
        openTimePicker: '选择时间',
        closeTimePicker: '返回日期',
        timePicker: '时间选择器',
        toggleOverlay: '切换选择器',
      }"
      :ui="{
        input: 'input',
        menu: timeEditorOpen ? 'sage-calendar calendar-editing-time' : 'sage-calendar',
      }"
      :teleport="portal || false"
      :transitions="{
        menuAppearTop: 'calendar-above',
        menuAppearBottom: 'calendar-below',
        open: 'calendar-overlay',
        close: 'calendar-overlay',
      }"
      :config="{ allowPreventDefault: true, onInternalKeydown: preventDialogCancel }"
      arrow-navigation
      placeholder="YYYY-MM-DD HH:mm"
      @open="open = true"
      @closed="
        open = false;
        timeEditorOpen = false;
      "
      @text-input="validateInput"
      @update:model-value="update"
      @internal-model-change="updateTimePreview"
      @overlay-toggle="updateOverlayState"
    >
      <template #input-icon>
        <button
          type="button"
          class="calendar-trigger"
          :aria-label="`选择${label}`"
          aria-haspopup="dialog"
          :aria-expanded="open"
          :disabled="disabled"
          @click.stop="picker?.toggleMenu()"
        >
          <PhCalendarBlank :size="19" />
        </button>
      </template>
      <template #clock-icon>
        <span ref="timePreview" class="calendar-time-preview">
          <PhClock :size="19" />
          <span v-if="selectedTime">{{ selectedTime }}</span>
        </span>
      </template>
      <template #time-picker-overlay="time">
        <CalendarTimeEditor
          :hours="time.hours as number"
          :minutes="time.minutes as number"
          :date="selectedDate"
          @confirm="confirmTime($event, time)"
          @cancel="closeTimeEditor"
        />
      </template>
    </VueDatePicker>
  </div>
</template>
