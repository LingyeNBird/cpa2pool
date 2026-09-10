<script setup lang="ts">
import { computed } from 'vue';
import './AnimatedValue.css';
const props = defineProps<{ value: string | number }>();
const text = computed(() => String(props.value));
</script>
<template>
  <span class="animated-value" role="img" :aria-label="text">
    <Transition name="number-swap" appear>
      <span :key="text" class="animated-value-face" aria-hidden="true">{{ text }}</span>
    </Transition>
  </span>
</template>
