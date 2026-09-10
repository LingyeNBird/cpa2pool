<script setup lang="ts">
import AnimatedValue from './components/AnimatedValue.vue';
import { computed, onMounted, ref } from 'vue';
import { PhPlus, PhPencilSimple, PhTrash } from '@phosphor-icons/vue';
import { api, act, busy, money } from '../api';
import type { Price } from '../types';
import PriceEditor from './components/PriceEditor.vue';
import DialogFrame from './components/DialogFrame.vue';
import FieldLabel from './components/FieldLabel.vue';
import { loadModelCatalog } from '../modelCatalog';
import './PricesPage.css';
const prices = ref<Price[]>([]);
const editing = ref<Price | null>(null);
const showEditor = ref(false);
const removing = ref<Price | null>(null);
const activeTable = ref<'token' | 'image' | 'video'>('token');
const editingMode = ref<'token' | 'image' | 'video'>('token');
const imagePrices = computed(() => prices.value.filter((price) => price.billing_mode === 'image'));
const videoPrices = computed(() =>
  prices.value.filter(
    (price) =>
      Number(price.video_price_480p) ||
      Number(price.video_price_720p) ||
      Number(price.video_price_1024p) ||
      Number(price.video_price_1080p),
  ),
);
const loading = ref(true);
const catalogError = ref('');
async function load() {
  loading.value = true;
  catalogError.value = '';
  try {
    prices.value = await api<Price[]>('prices');
    try {
      const catalog = await loadModelCatalog();
      prices.value = await api<Price[]>(
        'prices/sync-defaults',
        'POST',
        catalog.map((item) => item.price),
      );
    } catch (e) {
      catalogError.value = e instanceof Error ? e.message : String(e);
    }
  } finally {
    loading.value = false;
  }
}
async function saved() {
  showEditor.value = false;
  await act(load);
}
async function remove() {
  const p = removing.value;
  if (!p) return;
  await act(async () => {
    await api(`prices?model=${encodeURIComponent(p.model)}`, 'DELETE');
    removing.value = null;
    await load();
  });
}
onMounted(() => act(load));
</script>
<template>
  <section>
    <div class="section-heading">
      <h2>计费设置</h2>
      <button
        class="btn btn-primary"
        @click="
          editing = null;
          editingMode = 'token';
          showEditor = true;
        "
      >
        <PhPlus :size="18" />添加模型
      </button>
    </div>
    <div class="panel">
      <div class="pricing-tabs">
        <button
          class="btn btn-sm"
          :class="activeTable === 'token' ? 'btn-primary' : 'btn-ghost'"
          @click="activeTable = 'token'"
        >
          Token 计费
        </button>
        <button
          class="btn btn-sm"
          :class="activeTable === 'image' ? 'btn-primary' : 'btn-ghost'"
          @click="activeTable = 'image'"
        >
          图片生成计费
        </button>
        <button
          class="btn btn-sm"
          :class="activeTable === 'video' ? 'btn-primary' : 'btn-ghost'"
          @click="activeTable = 'video'"
        >
          视频生成计费
        </button>
      </div>
      <div v-if="catalogError" class="alert">{{ catalogError }}</div>
      <div v-if="loading" class="empty">读取中</div>
      <div v-else-if="!prices.length" class="empty">添加模型价格后可开始计费</div>
      <div v-else-if="activeTable === 'token'" class="table-wrap table-list table-list-inset">
        <table class="table">
          <thead>
            <tr>
              <th>模型</th>
              <th><FieldLabel text="输入" tip="USD / 百万 Token。" /></th>
              <th>输出</th>
              <th>缓存读取</th>
              <th>缓存写入</th>
              <th>倍率</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in prices" :key="p.model">
              <td><strong>{{ p.model }}</strong></td>
              <td class="amount"><AnimatedValue :value="money(p.input)" /></td>
              <td class="amount"><AnimatedValue :value="money(p.output)" /></td>
              <td class="amount"><AnimatedValue :value="money(p.cache_read)" /></td>
              <td class="amount"><AnimatedValue :value="money(p.cache_write)" /></td>
              <td>
                <div class="row-actions">
                  <span v-if="p.priority_enabled" class="badge"
                    >FAST ×<AnimatedValue :value="p.priority_multiplier"
                  /></span>
                  <span v-if="p.long_enabled" class="badge">长上下文</span>
                  <span v-if="p.model_enabled" class="badge"
                    >模型 ×<AnimatedValue :value="p.model_multiplier"
                  /></span>
                  <span v-if="!p.priority_enabled && !p.long_enabled && !p.model_enabled">无</span>
                </div>
              </td>
              <td>
                <div class="row-actions">
                  <button
                    class="btn btn-sm"
                    @click="
                      editing = p;
                      editingMode = 'token';
                      showEditor = true;
                    "
                  >
                    <PhPencilSimple :size="16" />编辑
                  </button>
                  <button
                    class="btn btn-sm btn-ghost btn-square"
                    :aria-label="`删除${p.model}`"
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
      <div
        v-else-if="activeTable === 'image' && imagePrices.length"
        class="table-wrap table-list table-list-inset"
      >
        <table class="table">
          <thead>
            <tr>
              <th>模型</th>
              <th>1K / 张</th>
              <th>2K / 张</th>
              <th>4K / 张</th>
              <th>模型倍率</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in imagePrices" :key="p.model">
              <td><strong>{{ p.model }}</strong></td>
              <td class="amount"><AnimatedValue :value="money(p.image_price_1k)" /></td>
              <td class="amount"><AnimatedValue :value="money(p.image_price_2k)" /></td>
              <td class="amount"><AnimatedValue :value="money(p.image_price_4k)" /></td>
              <td>
                <span v-if="p.model_enabled"
                  >×<AnimatedValue :value="p.model_multiplier"
                /></span>
                <span v-else>无</span>
              </td>
              <td>
                <button
                  class="btn btn-sm"
                  @click="
                    editing = p;
                    editingMode = 'image';
                    showEditor = true;
                  "
                >
                  <PhPencilSimple :size="16" />编辑
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div
        v-else-if="activeTable === 'video' && videoPrices.length"
        class="table-wrap table-list table-list-inset"
      >
        <table class="table">
          <thead>
            <tr>
              <th>模型</th>
              <th>480p / 秒</th>
              <th>720p / 秒</th>
              <th>1024p / 秒</th>
              <th>1080p / 秒</th>
              <th>模型倍率</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in videoPrices" :key="p.model">
              <td><strong>{{ p.model }}</strong></td>
              <td class="amount"><AnimatedValue :value="money(p.video_price_480p)" /></td>
              <td class="amount"><AnimatedValue :value="money(p.video_price_720p)" /></td>
              <td class="amount"><AnimatedValue :value="money(p.video_price_1024p)" /></td>
              <td class="amount"><AnimatedValue :value="money(p.video_price_1080p)" /></td>
              <td>
                <span v-if="p.model_enabled">×<AnimatedValue :value="p.model_multiplier" /></span>
                <span v-else>无</span>
              </td>
              <td>
                <button
                  class="btn btn-sm"
                  @click="
                    editing = p;
                    editingMode = 'video';
                    showEditor = true;
                  "
                >
                  <PhPencilSimple :size="16" />编辑
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="empty">
        当前模型目录没有{{ activeTable === 'video' ? '视频' : '图片' }}生成模型
      </div>
    </div>
    <PriceEditor
      v-if="showEditor"
      :price="editing"
      :mode="editingMode"
      @close="showEditor = false"
      @saved="saved"
    />
    <DialogFrame v-if="removing" title="删除模型价格" @close="removing = null">
      <p>删除「{{ removing.model }}」的价格？新请求将被拒绝，历史账单保留。</p>
      <div class="form-actions">
        <button class="btn" @click="removing = null">取消</button>
        <button class="btn btn-error" :disabled="busy" @click="remove">删除</button>
      </div>
    </DialogFrame>
  </section>
</template>
