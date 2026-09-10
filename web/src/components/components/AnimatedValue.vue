<script setup lang="ts">
import { computed } from 'vue';
import './AnimatedValue.css';
const props = defineProps<{ value: string | number }>();
const text = computed(() => String(props.value));
const places = computed(() => {
  const match = text.value.match(/^([^\d]*)([\d,]+)(\.\d+)?([^\d]*)$/);
  if (!match) return [{ key: 'literal', character: text.value, literal: true }];
  const [, prefix, integer, fraction = '', suffix] = match;
  const result: { key: string; character: string; literal?: boolean }[] = [];
  for (let i = 0; i < prefix.length; i++) result.push({ key: `prefix-${i}`, character: prefix[i] });
  let position = integer.replaceAll(',', '').length;
  for (const character of integer) {
    if (character === ',') result.push({ key: `group-${position}`, character });
    else result.push({ key: `integer-${--position}`, character });
  }
  for (let i = 0; i < fraction.length; i++) {
    result.push({ key: i === 0 ? 'decimal' : `fraction-${i}`, character: fraction[i] });
  }
  for (let i = 0; i < suffix.length; i++) result.push({ key: `suffix-${i}`, character: suffix[i] });
  return result;
});
</script>
<template>
  <TransitionGroup
    name="number-place"
    tag="span"
    class="animated-value"
    role="img"
    :aria-label="text"
  >
    <span
      v-for="place in places"
      :key="place.key"
      class="number-place"
      :class="{
        'number-punctuation': place.character === ',' || place.character === '.',
        'number-literal': place.literal,
      }"
      aria-hidden="true"
    >
      <span class="number-cell">
        <Transition name="number-swap">
          <span :key="place.character" class="animated-value-face">{{ place.character }}</span>
        </Transition>
      </span>
    </span>
  </TransitionGroup>
</template>
