<script setup lang="ts">
import type { Audit } from '../../types';
import { date } from '../../api';
import DialogFrame from './DialogFrame.vue';
import './AuditDetail.css';
defineProps<{ audit: Audit }>();
const emit = defineEmits<{ close: [] }>();
</script>
<template>
  <DialogFrame title="调整记录" @close="emit('close')"
    ><dl class="detail-list">
      <dt>时间</dt>
      <dd>{{ date(audit.time) }}</dd>
      <dt>操作</dt>
      <dd>{{ audit.action }}</dd>
      <dt>备注</dt>
      <dd>{{ audit.note || '无' }}</dd>
    </dl>
    <div class="audit-snapshots">
      <section>
        <h3>调整前</h3>
        <pre>{{ JSON.stringify(audit.before, null, 2) }}</pre>
      </section>
      <section>
        <h3>调整后</h3>
        <pre>{{ JSON.stringify(audit.after, null, 2) }}</pre>
      </section>
    </div></DialogFrame
  >
</template>
