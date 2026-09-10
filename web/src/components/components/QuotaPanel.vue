<script setup lang="ts">
import AnimatedValue from './AnimatedValue.vue';
import { onMounted, ref } from 'vue';
import { PhPlus } from '@phosphor-icons/vue';
import { api, act, money, date, reason, periods } from '../../api';
import type { Participant, Status, Quota, QuotaView } from '../../types';
import DialogFrame from './DialogFrame.vue';
import QuotaEditor from './components/QuotaEditor.vue';
import QuotaAdjustment from './components/QuotaAdjustment.vue';
import './QuotaPanel.css';
const props = defineProps<{ participant: Participant }>();
const emit = defineEmits<{ close: []; changed: [] }>();
const status = ref<Status>();
const editing = ref<Quota | null>(null);
const showEditor = ref(false);
const adjusting = ref<QuotaView | null>(null);
async function load() {
  status.value = await api<Status>(`status?participant_id=${props.participant.id}`);
}
async function saved() {
  showEditor.value = false;
  adjusting.value = null;
  await act(load);
  emit('changed');
}
onMounted(() => act(load));
</script>
<template>
  <DialogFrame :title="`${participant.name} · 额度`" @close="emit('close')">
    <dl class="detail-list">
      <dt>参与者标识</dt>
      <dd class="numeric">{{ participant.id }}</dd>
      <dt>当前状态</dt>
      <dd>{{ status?.available ? '可用' : status?.reasons.map(reason).join('、') }}</dd>
      <dt>累计消费</dt>
      <dd class="amount"><AnimatedValue :value="money(status?.used)" /></dd>
      <dt>可用额度</dt>
      <dd class="amount"><AnimatedValue :value="money(status?.remaining)" /></dd>
    </dl>
    <hr class="divider-line" />
    <div class="section-heading">
      <h3>额度限制</h3>
      <button
        class="btn btn-sm btn-primary"
        @click="
          editing = null;
          showEditor = true;
        "
      >
        <PhPlus :size="16" />添加额度
      </button>
    </div>
    <div v-if="!status?.quotas.length" class="empty">添加额度后可开始使用</div>
    <div v-for="q in status?.quotas" :key="q.id" class="quota-item">
      <div class="section-heading">
        <h3>
          {{ q.name }} <span class="badge">{{ periods[q.period] }}</span>
        </h3>
        <span v-if="q.reason" class="badge">{{ reason(q.reason) }}</span>
      </div>
      <div class="quota-amounts">
        <div>
          <span>限额</span><strong><AnimatedValue :value="money(q.current.limit)" /></strong>
        </div>
        <div>
          <span>已用</span><strong><AnimatedValue :value="money(q.current.used)" /></strong>
        </div>
        <div>
          <span>剩余</span><strong><AnimatedValue :value="money(q.current.remaining)" /></strong>
        </div>
      </div>
      <dl class="quota-dates">
        <dt>周期开始</dt>
        <dd>{{ date(q.current.starts_at) }}</dd>
        <dt>周期结束</dt>
        <dd>{{ date(q.current.ends_at) }}</dd>
        <dt>有效期</dt>
        <dd>{{ date(q.expires_at) }}</dd>
      </dl>
      <div class="form-actions">
        <button
          class="btn btn-sm"
          @click="
            editing = q;
            showEditor = true;
          "
        >
          编辑</button
        ><button class="btn btn-sm btn-primary" @click="adjusting = q">调整额度</button>
      </div>
    </div> </DialogFrame
  ><QuotaEditor
    v-if="showEditor"
    :quota="editing"
    :participant-id="participant.id"
    @close="showEditor = false"
    @saved="saved"
  /><QuotaAdjustment v-if="adjusting" :quota="adjusting" @close="adjusting = null" @saved="saved" />
</template>
