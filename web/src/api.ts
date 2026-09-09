import { ref } from 'vue';
const key = ref('');
export const authenticated = ref(false);
export const failure = ref('');
export const busy = ref(false);
export async function api<T>(path: string, method = 'GET', body?: unknown): Promise<T> {
  const options: RequestInit = {
    method,
    headers: { Authorization: `Bearer ${key.value}`, 'Content-Type': 'application/json' },
  };
  if (body !== undefined) options.body = JSON.stringify(body);
  const response = await fetch(`/v0/management/cpa2pool/${path}`, options);
  if (response.status === 401) authenticated.value = false;
  const data = await response.json();
  if (!response.ok) throw new Error(data.error || `请求失败 (${response.status})`);
  return data.data as T;
}
export async function login(value: string) {
  key.value = value;
  await api('participants');
  authenticated.value = true;
}
export function logout() {
  key.value = '';
  authenticated.value = false;
}
export async function act(fn: () => Promise<void>) {
  busy.value = true;
  failure.value = '';
  try {
    await fn();
  } catch (e) {
    failure.value = e instanceof Error ? e.message : String(e);
  } finally {
    busy.value = false;
  }
}
export function money(value: string | null | undefined) {
  return value == null
    ? '—'
    : `$${Number(value).toLocaleString('en-US', { maximumFractionDigits: 9 })}`;
}
export function date(value: string | null | undefined) {
  return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '长期';
}
export function localDate(value?: string | null) {
  if (!value) return '';
  const d = new Date(value);
  return new Date(d.getTime() - d.getTimezoneOffset() * 60000).toISOString().slice(0, 16);
}
export function iso(value: string) {
  return value ? new Date(value).toISOString() : null;
}
export const periods: Record<string, string> = {
  none: '长期',
  day: '每日',
  week: '每周',
  month: '每月',
};
export const reasons: Record<string, string> = {
  deleted: '已删除',
  disabled: '已停用',
  paused: '已暂停',
  expired: '已过期',
  no_active_quota: '无可用额度',
  exhausted: '额度用尽',
  not_started: '未开始',
};
export function reason(value: string) {
  return value.startsWith('quota_exhausted:') ? '额度用尽' : reasons[value] || value;
}
