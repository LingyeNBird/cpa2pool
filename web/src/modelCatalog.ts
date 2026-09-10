import { cpaManagement } from './api';

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

export interface ModelCatalogItem {
  id: string;
  input: string;
  output: string;
  cacheRead: string;
  cacheWrite: string;
  hasPrice: boolean;
}

const providerAliases: Record<string, string> = {
  claude: 'anthropic',
  gemini: 'google',
  grok: 'xai',
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

function decimal(value: number | undefined): string {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0 ? String(value) : '0';
}

export async function loadModelCatalog(): Promise<ModelCatalogItem[]> {
  const configured = await cpaManagement<{ 'api-keys'?: string[] }>('api-keys');
  const apiKey = configured['api-keys']?.find((value) => value.trim());
  if (!apiKey) throw new Error('CPA 尚未配置可用于读取模型目录的 API Key');

  const [cpaResponse, modelsDevResponse] = await Promise.all([
    fetch('/v1/models', { headers: { Authorization: `Bearer ${apiKey}` } }),
    fetch('https://models.dev/api.json'),
  ]);
  if (!cpaResponse.ok) throw new Error(`CPA 模型目录读取失败 (${cpaResponse.status})`);
  if (!modelsDevResponse.ok)
    throw new Error(`models.dev 价格读取失败 (${modelsDevResponse.status})`);

  const cpa = (await cpaResponse.json()) as CPAModelsResponse;
  const providers = (await modelsDevResponse.json()) as Record<string, ModelsDevProvider>;
  const supported = new Map<string, string>();
  for (const model of cpa.data || []) {
    if (typeof model.id !== 'string' || !model.id.trim()) continue;
    const id = model.id.trim();
    const owner = typeof model.owned_by === 'string' ? model.owned_by : '';
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
          return model ? [{ provider, model }] : [];
        })
        .sort((left, right) => left.provider.localeCompare(right.provider));
      const preferred = preferredProvider(id, owner);
      const match = matches.find(({ provider }) => provider === preferred) || matches[0];
      const cost = match?.model.cost;
      return {
        id,
        input: decimal(cost?.input),
        output: decimal(cost?.output),
        cacheRead: decimal(cost?.cache_read),
        cacheWrite: decimal(cost?.cache_write),
        hasPrice: Boolean(cost),
      };
    });
}
