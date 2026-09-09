<script setup lang="ts">
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
    <div class="table-wrap">
      <table class="table">
        <thead>
          <tr>
            <th>项目</th>
            <th>Token</th>
            <th>单价 / 百万</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>输入</td>
            <td>{{ bill.usage.input }}</td>
            <td>{{ money(bill.charge.price.input) }}</td>
          </tr>
          <tr>
            <td>输出</td>
            <td>{{ bill.usage.output }}</td>
            <td>{{ money(bill.charge.price.output) }}</td>
          </tr>
          <tr>
            <td>缓存读取</td>
            <td>{{ bill.usage.cache_read }}</td>
            <td>{{ money(bill.charge.price.cache_read) }}</td>
          </tr>
          <tr>
            <td>缓存写入</td>
            <td>{{ bill.usage.cache_write }}</td>
            <td>{{ money(bill.charge.price.cache_write) }}</td>
          </tr>
          <tr>
            <td>推理（已含输出）</td>
            <td>{{ bill.usage.reasoning }}</td>
            <td>不重复计费</td>
          </tr>
        </tbody>
      </table>
    </div>
    <hr class="divider-line" />
    <dl class="detail-list">
      <dt>基础费用</dt>
      <dd class="amount">{{ money(bill.charge.base) }}</dd>
      <template v-for="f in bill.charge.factors" :key="f.name"
        ><dt>{{ factors[f.name] }}</dt>
        <dd>输入 ×{{ f.input }} · 输出 ×{{ f.output }}</dd></template
      >
      <dt>组合方式</dt>
      <dd>{{ bill.charge.price.combination === 'max' ? '取最大值' : '相乘' }}</dd>
      <dt>最终扣费</dt>
      <dd class="amount">{{ money(bill.charge.final) }}</dd>
      <dt>价格版本时间</dt>
      <dd>{{ date(bill.charge.price.updated_at) }}</dd>
    </dl></DialogFrame
  >
</template>
