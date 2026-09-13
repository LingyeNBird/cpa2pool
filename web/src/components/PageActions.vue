<script setup lang="ts">
import { PhArrowClockwise, PhSignOut, PhChatCircleText } from '@phosphor-icons/vue';
import './PageActions.css';
defineProps<{ busy: boolean }>();
const emit = defineEmits<{ refresh: []; logout: []; feedback: [] }>();
</script>

<template>
  <div class="page-actions" role="group" aria-label="页面操作">
    <button class="btn btn-sm" :disabled="busy" @click="emit('refresh')">
      <PhArrowClockwise :size="18" />刷新
    </button>
    <button
      class="btn btn-sm btn-ghost"
      title="仅退出当前插件页面；刷新后会重新使用宿主登录。"
      @click="emit('logout')"
    >
      <PhSignOut :size="18" />退出
    </button>
    <button class="btn btn-sm btn-ghost" @click="emit('feedback')">
      <PhChatCircleText :size="18" />反馈
    </button>
  </div>
</template>
