import { cpaManagement } from './api';
import type { Price } from './types';

interface CPAModel {
  id?: unknown;
  owned_by?: unknown;
}
interface CPAModelsResponse {
  data?: CPAModel[];
}
interface ModelsDevCost {
  input?: number;
  output?: number;
  cache_read?: number;
  cache_write?: number;
}
interface ModelsDevModel {
  id?: string;
  cost?: ModelsDevCost;
}
interface ModelsDevProvider {
  models?: Record<string, ModelsDevModel>;
}
interface Sub2APIPrice {
  input_cost_per_token?: number;
  output_cost_per_token?: number;
  cache_read_input_token_cost?: number;
  cache_creation_input_token_cost?: number;
  input_cost_per_token_priority?: number;
  output_cost_per_token_priority?: number;
  long_context_input_token_threshold?: number;
  long_context_input_cost_multiplier?: number;
  long_context_output_cost_multiplier?: number;
  output_cost_per_image?: number;
  mode?: string;
}
export interface ModelCatalogItem {
  id: string;
  price: Price;
  source: 'models.dev' | 'sub2api' | 'sub2api-fallback';
}

const SUB2API_PRICES =
  'https://raw.githubusercontent.com/Wei-Shaw/model-price-repo/main/model_prices_and_context_window.json';
const providerAliases: Record<string, string> = {
  claude: 'anthropic',
  gemini: 'google',
  grok: 'xai',
};
const staticFallbacks: Record<string, Sub2APIPrice> = {
  'gpt-5.4': {
    input_cost_per_token: 2.5e-6,
    output_cost_per_token: 1.5e-5,
    cache_read_input_token_cost: 2.5e-7,
    long_context_input_token_threshold: 272000,
    long_context_input_cost_multiplier: 2,
    long_context_output_cost_multiplier: 1.5,
  },
  'gpt-5.4-mini': {
    input_cost_per_token: 7.5e-7,
    output_cost_per_token: 4.5e-6,
    cache_read_input_token_cost: 7.5e-8,
  },
  'gpt-5.4-nano': {
    input_cost_per_token: 2e-7,
    output_cost_per_token: 1.25e-6,
    cache_read_input_token_cost: 2e-8,
  },
  'gpt-5.6-sol': {
    input_cost_per_token: 5e-6,
    output_cost_per_token: 3e-5,
    cache_read_input_token_cost: 5e-7,
    cache_creation_input_token_cost: 6.25e-6,
    input_cost_per_token_priority: 1e-5,
    output_cost_per_token_priority: 6e-5,
    long_context_input_token_threshold: 272000,
    long_context_input_cost_multiplier: 2,
    long_context_output_cost_multiplier: 1.5,
  },
  'gpt-5.6-terra': {
    input_cost_per_token: 2e-6,
    output_cost_per_token: 1.2e-5,
    cache_read_input_token_cost: 2e-7,
    cache_creation_input_token_cost: 2.5e-6,
    input_cost_per_token_priority: 4e-6,
    output_cost_per_token_priority: 2.4e-5,
    long_context_input_token_threshold: 272000,
    long_context_input_cost_multiplier: 2,
    long_context_output_cost_multiplier: 1.5,
  },
  'gpt-5.6-luna': {
    input_cost_per_token: 2e-7,
    output_cost_per_token: 1.2e-6,
    cache_read_input_token_cost: 2e-8,
    cache_creation_input_token_cost: 2.5e-7,
    input_cost_per_token_priority: 4e-7,
    output_cost_per_token_priority: 2.4e-6,
    long_context_input_token_threshold: 272000,
    long_context_input_cost_multiplier: 2,
    long_context_output_cost_multiplier: 1.5,
  },
};

function canonicalProvider(model: string): string {
  if (/^(gpt-|chatgpt-|o[134]-|codex-)/i.test(model)) return 'openai';
  if (/^claude-/i.test(model)) return 'anthropic';
  if (/^gemini-/i.test(model)) return 'google';
  if (/^grok-/i.test(model)) return 'xai';
  return '';
}
function preferredProvider(model: string, ownedBy: string): string {
  const owner = ownedBy.trim().toLowerCase();
  return providerAliases[owner] || owner || canonicalProvider(model);
}
function million(value: number | undefined): number {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0
    ? Number((value * 1_000_000).toPrecision(12))
    : 0;
}
function ratio(priority: number | undefined, base: number | undefined): number | undefined {
  return priority !== undefined && base !== undefined && base > 0 ? priority / base : undefined;
}
function priceFromModelsDev(id: string, cost: ModelsDevCost): Price {
  return defaultPrice(id, {
    input: cost.input,
    output: cost.output,
    cacheRead: cost.cache_read,
    cacheWrite: cost.cache_write,
  });
}
function priceFromSub2API(id: string, value: Sub2APIPrice): Price {
  const effective = { ...value };
  const normalized = normalizeModel(id);
  const isGPT56 = normalized.startsWith('gpt-5.6');
  const usesLegacyLongContext =
    normalized === 'gpt-5.4' ||
    normalized.startsWith('gpt-5.4-') ||
    normalized === 'gpt-5.5' ||
    normalized.startsWith('gpt-5.5-');
  if (isGPT56 && effective.cache_creation_input_token_cost === undefined) {
    effective.cache_creation_input_token_cost = (effective.input_cost_per_token || 0) * 1.25;
  }
  if (isGPT56 || usesLegacyLongContext) {
    effective.long_context_input_token_threshold ||= 272000;
    effective.long_context_input_cost_multiplier ||= 2;
    effective.long_context_output_cost_multiplier ||= 1.5;
  }
  const inputRatio = ratio(effective.input_cost_per_token_priority, effective.input_cost_per_token);
  const outputRatio = ratio(
    effective.output_cost_per_token_priority,
    effective.output_cost_per_token,
  );
  const priorityMultiplier =
    inputRatio !== undefined && outputRatio !== undefined && inputRatio === outputRatio
      ? inputRatio
      : undefined;
  return defaultPrice(id, {
    input: million(effective.input_cost_per_token),
    output: million(effective.output_cost_per_token),
    cacheRead: million(effective.cache_read_input_token_cost),
    cacheWrite: million(effective.cache_creation_input_token_cost),
    priorityMultiplier,
    longThreshold: effective.long_context_input_token_threshold,
    longInputMultiplier: effective.long_context_input_cost_multiplier,
    longOutputMultiplier: effective.long_context_output_cost_multiplier,
  });
}
function imageModel(model: string, pricing?: Sub2APIPrice): boolean {
  const normalized = normalizeModel(model);
  return (
    pricing?.mode === 'image_generation' ||
    normalized.startsWith('gpt-image-') ||
    (normalized.startsWith('grok-imagine') && !normalized.includes('video')) ||
    (normalized.startsWith('gemini-') && normalized.includes('-image'))
  );
}
function withImagePricing(model: string, price: Price, pricing?: Sub2APIPrice): Price {
  if (!imageModel(model, pricing)) return price;
  const normalized = normalizeModel(model);
  let price1K = pricing?.output_cost_per_image || 0.134;
  let price2K = price1K * 1.5;
  let price4K = price1K * 2;
  if (
    normalized === 'grok-imagine' ||
    normalized === 'grok-imagine-image' ||
    normalized === 'grok-imagine-edit'
  ) {
    price1K = 0.02;
    price2K = 0.02;
    price4K = 0.02;
  } else if (normalized === 'grok-imagine-image-quality') {
    price1K = 0.05;
    price2K = 0.07;
    price4K = 0.07;
  }
  return {
    ...price,
    billing_mode: 'image',
    image_price_1k: String(price1K),
    image_price_2k: String(price2K),
    image_price_4k: String(price4K),
  };
}

function withVideoPricing(model: string, price: Price): Price {
  const normalized = normalizeModel(model);
  const rates: Record<string, [number, number, number, number]> = {
    'sora-2': [0, 0.1, 0, 0],
    'sora-2-pro': [0, 0.3, 0.5, 0.7],
    'grok-imagine-video': [0.05, 0.05, 0.05, 0.05],
    'grok-imagine-video-1.5': [0.08, 0.08, 0.08, 0.08],
    'grok-imagine-video-1.5-preview': [0.08, 0.08, 0.08, 0.08],
  };
  const rate = rates[normalized];
  if (!rate) return price;
  return {
    ...price,
    video_price_480p: String(rate[0]),
    video_price_720p: String(rate[1]),
    video_price_1024p: String(rate[2]),
    video_price_1080p: String(rate[3]),
  };
}

function defaultPrice(
  model: string,
  values: {
    input?: number;
    output?: number;
    cacheRead?: number;
    cacheWrite?: number;
    priorityMultiplier?: number;
    longThreshold?: number;
    longInputMultiplier?: number;
    longOutputMultiplier?: number;
  },
): Price {
  const longEnabled =
    values.longThreshold !== undefined &&
    values.longInputMultiplier !== undefined &&
    values.longOutputMultiplier !== undefined;
  return {
    billing_mode: 'token',
    image_price_1k: '0',
    image_price_2k: '0',
    image_price_4k: '0',
    video_price_480p: '0',
    video_price_720p: '0',
    video_price_1024p: '0',
    video_price_1080p: '0',
    model,
    input: String(values.input ?? 0),
    output: String(values.output ?? 0),
    cache_read: String(values.cacheRead ?? 0),
    cache_write: String(values.cacheWrite ?? 0),
    priority_enabled: values.priorityMultiplier !== undefined,
    priority_multiplier: String(values.priorityMultiplier ?? 2),
    long_enabled: longEnabled,
    long_threshold: values.longThreshold ?? 200000,
    long_input_multiplier: String(values.longInputMultiplier ?? 2),
    long_output_multiplier: String(values.longOutputMultiplier ?? 2),
    model_enabled: false,
    model_multiplier: '1',
    combination: 'multiply',
    updated_at: '',
  };
}
function normalizeModel(model: string): string {
  const lower = model.trim().toLowerCase().replace(/^\/+/, '');
  const marker = '/publishers/google/models/';
  if (lower.includes(marker)) return lower.slice(lower.lastIndexOf(marker) + marker.length);
  if (lower.startsWith('publishers/google/models/')) return lower.slice(25);
  if (lower.includes('/models/')) return lower.slice(lower.lastIndexOf('/models/') + 8);
  if (lower.startsWith('models/')) return lower.slice(7);
  return lower.includes('/') ? lower.slice(lower.lastIndexOf('/') + 1) : lower;
}
function withoutDate(model: string): string {
  return model.replace(/-\d{8}$/, '');
}
function findContaining(
  prices: Record<string, Sub2APIPrice>,
  patterns: readonly string[],
): Sub2APIPrice | undefined {
  for (const pattern of patterns) {
    const key = Object.keys(prices).find((candidate) => candidate.includes(pattern));
    if (key) return prices[key];
  }
}
function claudeFallback(
  model: string,
  prices: Record<string, Sub2APIPrice>,
): Sub2APIPrice | undefined {
  const families: Array<{ match: string[]; pricing?: string[] }> = [
    { match: ['claude-opus-5'], pricing: ['claude-opus-5', 'claude-opus-4-8'] },
    {
      match: ['claude-opus-4-8', 'claude-opus-4.8'],
      pricing: ['claude-opus-4-8', 'claude-opus-4.8', 'claude-opus-4-7'],
    },
    {
      match: ['claude-opus-4-7', 'claude-opus-4.7'],
      pricing: ['claude-opus-4-7', 'claude-opus-4.7', 'claude-opus-4-6'],
    },
    { match: ['claude-opus-4-6', 'claude-opus-4.6'] },
    { match: ['claude-opus-4-5', 'claude-opus-4.5'] },
    { match: ['claude-opus-4', 'claude-3-opus'] },
    { match: ['claude-sonnet-4-5', 'claude-sonnet-4.5'] },
    { match: ['claude-sonnet-4', 'claude-3-5-sonnet'] },
    { match: ['claude-3-5-sonnet', 'claude-3.5-sonnet'] },
    { match: ['claude-3-sonnet'] },
    { match: ['claude-3-5-haiku', 'claude-3.5-haiku'] },
    { match: ['claude-3-haiku'] },
  ];
  const family = families.find(({ match }) =>
    match.some((pattern) => model.includes(pattern) || model.includes(pattern.replaceAll('-', ''))),
  );
  return family ? findContaining(prices, family.pricing || family.match) : undefined;
}
function sub2APIFallback(
  modelID: string,
  source: Record<string, Sub2APIPrice>,
): { value: Sub2APIPrice; exact: boolean } | undefined {
  const prices = Object.fromEntries(
    Object.entries(source).map(([key, value]) => [key.toLowerCase(), value]),
  );
  const model = normalizeModel(modelID);
  const candidates = [model, model.replaceAll('-4-5-', '-4.5-'), withoutDate(model)];
  for (const candidate of candidates) {
    if (prices[candidate]) return { value: prices[candidate], exact: candidate === model };
  }
  const base = withoutDate(model);
  const sameBase = Object.entries(prices).find(([key]) => withoutDate(key) === base)?.[1];
  if (sameBase) return { value: sameBase, exact: false };
  const claude = claudeFallback(model, prices);
  if (claude) return { value: claude, exact: false };
  if (!model.startsWith('gpt-')) return undefined;

  const baseVersion = model.match(/^(gpt-\d+(?:\.\d+)?)(?:-|$)/)?.[1];
  if (model.startsWith('gpt-5.3-codex-spark') && prices['gpt-5.1-codex'])
    return { value: prices['gpt-5.1-codex'], exact: false };
  if (baseVersion && !model.startsWith('gpt-5.6-') && prices[baseVersion])
    return { value: prices[baseVersion], exact: false };
  if (model.startsWith('gpt-5.3-codex') && prices['gpt-5.2-codex'])
    return { value: prices['gpt-5.2-codex'], exact: false };
  for (const family of ['gpt-5.6-sol', 'gpt-5.6-terra', 'gpt-5.6-luna']) {
    if (model.startsWith(family))
      return { value: prices[family] || staticFallbacks[family], exact: false };
  }
  if (model.startsWith('gpt-5.5'))
    return { value: prices['gpt-5.4'] || staticFallbacks['gpt-5.4'], exact: false };
  for (const family of ['gpt-5.4-mini', 'gpt-5.4-nano', 'gpt-5.4']) {
    if (model.startsWith(family))
      return { value: prices[family] || staticFallbacks[family], exact: false };
  }
  if (model.startsWith('gpt-image-')) {
    for (const candidate of ['gpt-image-2', 'gpt-image-1.5', 'gpt-image-1']) {
      if (prices[candidate]) return { value: prices[candidate], exact: false };
    }
  }
  if (prices['gpt-5.1-codex']) return { value: prices['gpt-5.1-codex'], exact: false };
  return undefined;
}

export async function loadModelCatalog(): Promise<ModelCatalogItem[]> {
  const configured = await cpaManagement<{ 'api-keys'?: string[] }>('api-keys');
  const apiKey = configured['api-keys']?.find((value) => value.trim());
  if (!apiKey) throw new Error('CPA 尚未配置可用于读取模型目录的 API Key');

  const [cpaResponse, modelsDevResponse, sub2APIResponse] = await Promise.all([
    fetch('/v1/models', { headers: { Authorization: `Bearer ${apiKey}` } }),
    fetch('https://models.dev/api.json'),
    fetch(SUB2API_PRICES).catch(() => null),
  ]);
  if (!cpaResponse.ok) throw new Error(`CPA 模型目录读取失败 (${cpaResponse.status})`);
  if (!modelsDevResponse.ok)
    throw new Error(`models.dev 价格读取失败 (${modelsDevResponse.status})`);

  const cpa = (await cpaResponse.json()) as CPAModelsResponse;
  const providers = (await modelsDevResponse.json()) as Record<string, ModelsDevProvider>;
  const sub2Prices =
    sub2APIResponse?.ok === true
      ? ((await sub2APIResponse.json()) as Record<string, Sub2APIPrice>)
      : {};
  const supported = new Map<string, string>();
  for (const model of cpa.data || []) {
    if (typeof model.id !== 'string' || !model.id.trim()) continue;
    const id = model.id.trim();
    const owner = typeof model.owned_by === 'string' ? model.owned_by : '';
    if (!supported.has(id)) supported.set(id, owner);
  }
  for (const [id, owner] of [
    ['sora-2', 'openai'],
    ['sora-2-pro', 'openai'],
    ['grok-imagine-video', 'xai'],
    ['grok-imagine-video-1.5', 'xai'],
    ['grok-imagine-video-1.5-preview', 'xai'],
  ] as const) {
    if (!supported.has(id)) supported.set(id, owner);
  }

  return [...supported]
    .sort(([left], [right]) => left.localeCompare(right, undefined, { sensitivity: 'base' }))
    .map(([id, owner]) => {
      const matches = Object.entries(providers)
        .flatMap(([provider, definition]) => {
          const model = Object.values(definition.models || {}).find(
            (candidate) => candidate.id === id,
          );
          return model?.cost ? [{ provider, cost: model.cost }] : [];
        })
        .sort((left, right) => left.provider.localeCompare(right.provider));
      const preferred = preferredProvider(id, owner);
      const exact = matches.find(({ provider }) => provider === preferred) || matches[0];
      const fallback = sub2APIFallback(id, sub2Prices);
      if (exact)
        return {
          id,
          price: withVideoPricing(id, withImagePricing(id, priceFromModelsDev(id, exact.cost), fallback?.value)),
          source: 'models.dev' as const,
        };
      if (fallback)
        return {
          id,
          price: withVideoPricing(id, withImagePricing(id, priceFromSub2API(id, fallback.value), fallback.value)),
          source: fallback.exact ? ('sub2api' as const) : ('sub2api-fallback' as const),
        };
      return {
        id,
        price: withVideoPricing(id, withImagePricing(id, defaultPrice(id, {}))),
        source: 'sub2api-fallback' as const,
      };
    });
}
