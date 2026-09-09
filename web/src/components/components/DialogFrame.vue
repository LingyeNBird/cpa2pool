<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { PhX } from '@phosphor-icons/vue';
import './DialogFrame.css';
defineProps<{ title: string }>();
const emit = defineEmits<{ close: [] }>();
const dialog = ref<HTMLDialogElement>();
onMounted(() => dialog.value?.showModal());
</script>
<template><dialog ref="dialog" class="modal dialog-frame" @cancel.prevent="emit('close')" @click="($event.target === dialog) && emit('close')"><section class="modal-box panel"><header class="section-heading"><h2>{{ title }}</h2><button type="button" class="btn btn-ghost btn-square" aria-label="关闭对话框" @click="emit('close')"><PhX :size="20" /></button></header><slot /></section></dialog></template>
