<script setup lang="ts">
import AnimatedValue from './components/AnimatedValue.vue';
import { onMounted, ref, computed, watch } from 'vue';
import { PhPlus, PhPencilSimple, PhPause, PhPlay, PhTrash, PhWallet } from '@phosphor-icons/vue';
import { api, act, busy, money, reason, date } from '../api';
import type { Participant, Status } from '../types';
import ParticipantEditor from './components/ParticipantEditor.vue';
import QuotaPanel from './components/QuotaPanel.vue';
import DialogFrame from './components/DialogFrame.vue';
import TablePagination from './components/TablePagination.vue';
import './ParticipantsPage.css';
const participants = ref<Participant[]>([]);
const statuses = ref<Record<string, Status>>({});
const search = ref('');
const loading = ref(true);
const editing = ref<Participant | null>(null);
const showEditor = ref(false);
const selected = ref<Participant | null>(null);
const removing = ref<Participant | null>(null);
const shown = computed(() =>
  participants.value.filter((p) => `${p.name} ${p.note}`.includes(search.value)),
);
const page = ref(1);
const paged = computed(() => shown.value.slice((page.value - 1) * 10, page.value * 10));
watch(search, () => (page.value = 1));
watch(
  () => shown.value.length,
  (total) => (page.value = Math.min(page.value, Math.max(1, Math.ceil(total / 10)))),
);
async function load() {
  loading.value = true;
  try {
    participants.value = await api<Participant[]>('participants');
    const entries = await Promise.all(
      participants.value.map(
        async (p) => [p.id, await api<Status>(`status?participant_id=${p.id}`)] as const,
      ),
    );
    statuses.value = Object.fromEntries(entries);
  } finally {
    loading.value = false;
  }
}
async function saved() {
  showEditor.value = false;
  await act(load);
}
async function pause(p: Participant) {
  await act(async () => {
    await api('participants', 'PUT', { ...p, paused: !p.paused });
    await load();
  });
}
async function remove() {
  const p = removing.value;
  if (!p) return;
  await act(async () => {
    await api(`participants?id=${p.id}`, 'DELETE');
    removing.value = null;
    if (selected.value?.id === p.id) selected.value = null;
    await load();
  });
}
onMounted(() => act(load));
</script>
<template>
  <section>
    <div class="section-heading">
      <h2>
        参与者 <span class="participant-count"><AnimatedValue :value="participants.length" /></span>
      </h2>
      <button
        class="btn btn-primary"
        @click="
          editing = null;
          showEditor = true;
        "
      >
        <PhPlus :size="18" />添加参与者
      </button>
    </div>
    <div class="panel">
      <div class="participant-search">
        <input v-model="search" class="input" placeholder="搜索参与者" aria-label="搜索参与者" />
      </div>
      <div v-if="loading" class="empty">读取中</div>
      <div v-else-if="!shown.length" class="empty">
        {{ search ? '没有匹配的参与者' : '添加第一位参与者' }}
      </div>
      <div v-else class="table-wrap table-list table-list-inset">
        <table class="table">
          <thead>
            <tr>
              <th>参与者</th>
              <th>状态</th>
              <th>已用</th>
              <th>剩余</th>
              <th>有效期</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in paged" :key="p.id">
              <td>
                <button class="participant-name" @click="selected = p">{{ p.name }}</button>
                <div class="participant-key">{{ p.key_preview }}</div>
              </td>
              <td>
                <span
                  class="badge"
                  :class="statuses[p.id]?.available ? 'badge-success' : 'badge-ghost'"
                  >{{
                    statuses[p.id]?.available ? '可用' : reason(statuses[p.id]?.reasons[0] || '')
                  }}</span
                >
              </td>
              <td class="amount"><AnimatedValue :value="money(statuses[p.id]?.used)" /></td>
              <td class="amount"><AnimatedValue :value="money(statuses[p.id]?.remaining)" /></td>
              <td>{{ date(p.expires_at) }}</td>
              <td>
                <div class="row-actions">
                  <button class="btn btn-sm" @click="selected = p">
                    <PhWallet :size="16" />额度</button
                  ><button
                    class="btn btn-ghost btn-square btn-sm"
                    :aria-label="`编辑${p.name}`"
                    @click="
                      editing = p;
                      showEditor = true;
                    "
                  >
                    <PhPencilSimple :size="18" /></button
                  ><button
                    class="btn btn-ghost btn-square btn-sm"
                    :disabled="busy"
                    :aria-label="p.paused ? `恢复${p.name}` : `暂停${p.name}`"
                    @click="pause(p)"
                  >
                    <PhPlay v-if="p.paused" :size="18" /><PhPause v-else :size="18" /></button
                  ><button
                    class="btn btn-ghost btn-square btn-sm"
                    :aria-label="`删除${p.name}`"
                    @click="removing = p"
                  >
                    <PhTrash :size="18" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <TablePagination
        v-if="shown.length"
        :total="shown.length"
        :page="page"
        @page="page = $event"
      />
    </div>
    <ParticipantEditor
      v-if="showEditor"
      :participant="editing"
      @close="showEditor = false"
      @saved="saved"
    />
    <QuotaPanel
      v-if="selected"
      :participant="selected"
      @close="selected = null"
      @changed="act(load)"
    />
    <DialogFrame v-if="removing" title="删除参与者" @close="removing = null"
      ><p>删除「{{ removing.name }}」并停止使用？历史账单保留。</p>
      <div class="form-actions">
        <button class="btn" @click="removing = null">取消</button
        ><button class="btn btn-error" :disabled="busy" @click="remove">删除</button>
      </div></DialogFrame
    >
  </section>
</template>
