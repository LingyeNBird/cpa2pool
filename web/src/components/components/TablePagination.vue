<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import AnimatedValue from './AnimatedValue.vue';
import './TablePagination.css';

const props = defineProps<{ total: number; page: number; pageSize?: number; disabled?: boolean }>();
const emit = defineEmits<{ page: [value: number] }>();
const pages = computed(() => Math.max(1, Math.ceil(props.total / (props.pageSize || 10))));
const draft = ref(String(props.page));
watch(
  () => props.page,
  (value) => (draft.value = String(value)),
);
function jump() {
  const parsed = Number.parseInt(draft.value, 10);
  const target = Number.isFinite(parsed) ? Math.min(pages.value, Math.max(1, parsed)) : props.page;
  draft.value = String(target);
  if (target !== props.page) emit('page', target);
}
</script>

<template>
  <div class="table-pagination">
    <span class="table-pagination-count">共 <AnimatedValue :value="total" /> 条</span>
    <div class="table-pagination-controls">
      <span>共 <AnimatedValue :value="pages" /> 页</span>
      <button class="btn btn-sm" :disabled="disabled || page <= 1" @click="emit('page', page - 1)">
        上一页
      </button>
      <input
        v-model="draft"
        class="input input-sm table-page-input"
        type="number"
        min="1"
        :max="pages"
        aria-label="跳转页码"
        @keydown.enter.prevent="jump"
        @blur="draft = String(page)"
      />
      <button
        class="btn btn-sm"
        :disabled="disabled || page >= pages"
        @click="emit('page', page + 1)"
      >
        下一页
      </button>
    </div>
  </div>
</template>
