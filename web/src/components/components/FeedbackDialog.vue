<script setup lang="ts">
import { onBeforeUnmount, ref } from 'vue';
import { PhCheckCircle } from '@phosphor-icons/vue';
import { submitFeedback, FEEDBACK_LIMIT } from '../../api';
import DialogFrame from './DialogFrame.vue';
import './FeedbackDialog.css';
const emit = defineEmits<{ close: [] }>();
const content = ref('');
const sending = ref(false);
const succeeded = ref(false);
const error = ref('');
let closeTimer: ReturnType<typeof setTimeout> | undefined;
async function send() {
  const value = content.value.trim();
  if (!value) return;
  sending.value = true;
  error.value = '';
  try {
    await submitFeedback(value);
    succeeded.value = true;
    closeTimer = setTimeout(() => emit('close'), 2000);
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e);
  } finally {
    sending.value = false;
  }
}
onBeforeUnmount(() => clearTimeout(closeTimer));
</script>
<template>
  <DialogFrame title="意见反馈" @close="emit('close')">
    <div v-if="succeeded" class="feedback-success">
      <PhCheckCircle :size="56" weight="fill" />
      <p>反馈成功</p>
    </div>
    <form v-else @submit.prevent="send">
      <div v-if="error" class="alert alert-error error-bar">{{ error }}</div>
      <label class="field full-width"
        ><span>反馈内容</span
        ><textarea
          v-model="content"
          class="textarea"
          rows="5"
          :maxlength="FEEDBACK_LIMIT"
          placeholder="请输入您想反馈的内容"
          autofocus
      /></label>
      <div class="feedback-counter">{{ content.length }} / {{ FEEDBACK_LIMIT }}</div>
      <div class="form-actions">
        <button type="button" class="btn" @click="emit('close')">取消</button
        ><button class="btn btn-primary" :disabled="sending || !content.trim()">
          {{ sending ? '发送中' : '发送' }}
        </button>
      </div>
    </form>
  </DialogFrame>
</template>
