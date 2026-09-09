<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { PhPlus, PhPencilSimple, PhTrash } from '@phosphor-icons/vue';
import { api, act, busy, money } from '../api';
import type { Price } from '../types';
import PriceEditor from './components/PriceEditor.vue';
import DialogFrame from './components/DialogFrame.vue';
import FieldLabel from './components/FieldLabel.vue';
const prices = ref<Price[]>([]); const editing = ref<Price | null>(null); const showEditor = ref(false); const removing = ref<Price | null>(null); const loading = ref(true);
async function load() { loading.value = true; try { prices.value = await api<Price[]>('prices'); } finally { loading.value = false; } }
async function saved() { showEditor.value = false; await act(load); }
async function remove() { const p = removing.value; if (!p) return; await act(async () => { await api(`prices?model=${encodeURIComponent(p.model)}`, 'DELETE'); removing.value = null; await load(); }); }
onMounted(() => act(load));
</script>
<template><section><div class="section-heading"><h2>计费设置</h2><button class="btn btn-primary" @click="editing = null; showEditor = true"><PhPlus :size="18" />添加模型</button></div>
<div class="panel"><div v-if="loading" class="empty">读取中</div><div v-else-if="!prices.length" class="empty">添加模型价格后可开始计费</div><div v-else class="table-wrap"><table class="table"><thead><tr><th>模型</th><th><FieldLabel text="输入" tip="USD / 百万 Token" /></th><th>输出</th><th>缓存读取</th><th>缓存写入</th><th>倍率</th><th>操作</th></tr></thead><tbody><tr v-for="p in prices" :key="p.model"><td><strong>{{ p.model }}</strong></td><td class="amount">{{ money(p.input) }}</td><td class="amount">{{ money(p.output) }}</td><td class="amount">{{ money(p.cache_read) }}</td><td class="amount">{{ money(p.cache_write) }}</td><td><div class="row-actions"><span v-if="p.priority_enabled" class="badge">FAST ×{{ p.priority_multiplier }}</span><span v-if="p.long_enabled" class="badge">长上下文</span><span v-if="p.model_enabled" class="badge">模型 ×{{ p.model_multiplier }}</span><span v-if="!p.priority_enabled && !p.long_enabled && !p.model_enabled">无</span></div></td><td><div class="row-actions"><button class="btn btn-sm" @click="editing = p; showEditor = true"><PhPencilSimple :size="16" />编辑</button><button class="btn btn-sm btn-ghost btn-square" :aria-label="`删除${p.model}`" @click="removing = p"><PhTrash :size="18" /></button></div></td></tr></tbody></table></div></div>
<PriceEditor v-if="showEditor" :price="editing" @close="showEditor = false" @saved="saved" /><DialogFrame v-if="removing" title="删除模型价格" @close="removing = null"><p>删除「{{ removing.model }}」的价格？新请求将被拒绝，历史账单保留。</p><div class="form-actions"><button class="btn" @click="removing = null">取消</button><button class="btn btn-error" :disabled="busy" @click="remove">删除</button></div></DialogFrame></section></template>
