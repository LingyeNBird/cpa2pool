<script setup lang="ts">
import TooltipIcon from './TooltipIcon.vue';
import './FieldLabel.css';
defineProps<{ text: string; tip?: string }>();
</script>
<template>
  <span class="field-label">
    {{ text }}
    <TooltipIcon v-if="tip" :tip="tip" />
  </span>
</template>
