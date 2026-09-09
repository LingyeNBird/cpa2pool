<script setup lang="ts">
import { PhQuestion } from '@phosphor-icons/vue';
import './FieldLabel.css';
defineProps<{ text: string; tip?: string }>();
</script>
<template><span class="field-label">{{ text }}<span v-if="tip" class="tooltip tooltip-bottom" :data-tip="tip"><PhQuestion :size="16" tabindex="0" :aria-label="tip" /></span></span></template>
