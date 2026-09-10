<script setup lang="ts">
import AnimatedValue from './AnimatedValue.vue';
import type { Bill } from '../../types';
import { money, date } from '../../api';
import DialogFrame from './DialogFrame.vue';
defineProps<{ bill: Bill }>();
const emit = defineEmits<{ close: [] }>();
const factors: Record<string, string> = {
  priority: 'FAST / priority',
  long_context: '长上下文',
  model: '模型倍率',
};
</script>
<template>
  <DialogFrame title="请求账单" @close="emit('close')"
    ><dl class="detail-list">
      <dt>请求标识</dt>
      <dd>{{ bill.request_id }}</dd>
      <dt>时间</dt>
      <dd>{{ date(bill.time) }}</dd>
      <dt>实际模型</dt>
      <dd>{{ bill.model }}</dd>
      <dt>请求模型</dt>
      <dd>{{ bill.requested_model }}</dd>
      <dt>推理强度</dt>
      <dd>{{ bill.effort }}</dd>
      <dt>服务档位</dt>
      <dd>{{ bill.service_tier || '标准' }}</dd>
    </dl>
    <hr class="divider-line" />
    <div class="table-wrap table-detail">
      <table class="table">
        <thead>
          <tr>
            <th>项目</th>
            <th>数量</th>
            <th>单价</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>输入 Token</td>
            <td><AnimatedValue :value="bill.usage.input" /></td>
            <td><AnimatedValue :value="money(bill.charge.price.input)" /> / 百万</td>
          </tr>
          <tr>
            <td>输出 Token</td>
            <td><AnimatedValue :value="bill.usage.output" /></td>
            <td><AnimatedValue :value="money(bill.charge.price.output)" /> / 百万</td>
          </tr>
          <tr>
            <td>缓存读取</td>
            <td><AnimatedValue :value="bill.usage.cache_read" /></td>
            <td><AnimatedValue :value="money(bill.charge.price.cache_read)" /> / 百万</td>
          </tr>
          <tr>
            <td>缓存写入</td>
            <td><AnimatedValue :value="bill.usage.cache_write" /></td>
            <td><AnimatedValue :value="money(bill.charge.price.cache_write)" /> / 百万</td>
          </tr>
          <tr>
            <td>推理（已含输出）</td>
            <td><AnimatedValue :value="bill.usage.reasoning" /></td>
            <td>不重复计费</td>
          </tr>
          <tr v-if="bill.usage.images">
            <td>生成图片（{{ bill.usage.image_size }}）</td>
            <td><AnimatedValue :value="bill.usage.images" /></td>
            <td>
              <AnimatedValue
                :value="
                  money(
                    bill.usage.image_size === '1K'
                      ? bill.charge.price.image_price_1k
                      : bill.usage.image_size === '4K'
                        ? bill.charge.price.image_price_4k
                        : bill.charge.price.image_price_2k,
                  )
                "
              />
              / 张
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <hr class="divider-line" />
    <dl class="detail-list">
      <dt>基础费用</dt>
      <dd class="amount"><AnimatedValue :value="money(bill.charge.base)" /></dd>
      <template v-for="f in bill.charge.factors" :key="f.name"
        ><dt>{{ factors[f.name] }}</dt>
        <dd>输入 ×<AnimatedValue :value="f.input" /> · 输出 ×<AnimatedValue :value="f.output" /></dd
      ></template>
      <dt>组合方式</dt>
      <dd>{{ bill.charge.price.combination === 'max' ? '取最大值' : '相乘' }}</dd>
      <dt>最终扣费</dt>
      <dd class="amount"><AnimatedValue :value="money(bill.charge.final)" /></dd>
      <dt>价格版本时间</dt>
      <dd>{{ date(bill.charge.price.updated_at) }}</dd>
    </dl></DialogFrame
  >
</template>
