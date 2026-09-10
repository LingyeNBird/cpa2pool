<script setup lang="ts">
import AnimatedValue from './AnimatedValue.vue';
import type { Stat } from '../../types';
import { money } from '../../api';
import './SummaryStats.css';
defineProps<{ total?: Stat }>();
</script>
<template>
  <div class="summary-stats">
    <div>
      <span>实际消费</span><strong><AnimatedValue :value="money(total?.cost || '0')" /></strong>
    </div>
    <div>
      <span>请求数</span
      ><strong><AnimatedValue :value="(total?.requests || 0).toLocaleString()" /></strong>
    </div>
    <div>
      <span>输入 / 输出 Token</span
      ><strong
        ><AnimatedValue :value="(total?.input || 0).toLocaleString()" />
        <small>/ <AnimatedValue :value="(total?.output || 0).toLocaleString()" /></small
      ></strong>
    </div>
  </div>
</template>
