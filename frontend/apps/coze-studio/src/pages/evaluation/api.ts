/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import {
  type EvalSet,
  type EvalSetItem,
  type Evaluator,
  type EvaluatorTemplate,
  type Experiment,
  type ExperimentDetail,
  type ModelOption,
  type Paged,
  type ParsedImportFile,
  type TargetInfo,
} from './types';

const BASE = '/api/evaluation_api';

interface RequestOptions {
  method?: 'GET' | 'POST';
  params?: Record<string, unknown>;
  body?: unknown;
}

async function request<T>(path: string, options?: RequestOptions): Promise<T> {
  const method = options?.method ?? 'GET';
  let url = `${BASE}${path}`;
  if (options?.params) {
    const sp = new URLSearchParams();
    Object.entries(options.params).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        sp.set(key, String(value));
      }
    });
    const qs = sp.toString();
    if (qs) {
      url += `?${qs}`;
    }
  }
  const body = options?.body !== undefined ? JSON.stringify(options.body) : undefined;
  const resp = await fetch(url, {
    method,
    headers: body ? { 'Content-Type': 'application/json' } : undefined,
    body,
    credentials: 'same-origin',
  });
  let data: unknown = null;
  try {
    data = await resp.json();
  } catch {
    // non-json body, ignore
  }
  if (!resp.ok) {
    const error = new Error(
      (data as { msg?: string })?.msg || `HTTP ${resp.status}`,
    ) as Error & { code?: number };
    error.code = (data as { code?: number })?.code;
    throw error;
  }
  const body = data as { code?: number; msg?: string } | null;
  if (body && typeof body.code === 'number' && body.code !== 0) {
    const error = new Error(body.msg || '请求失败，请稍后重试') as Error & {
      code?: number;
    };
    error.code = body.code;
    throw error;
  }
  return data as T;
}

interface SuccessResp {
  success?: boolean;
}

export function getErrorMessage(e: unknown): string {
  return e instanceof Error ? e.message : '请求失败，请稍后重试';
}

export const evaluationApi = {
  createEvalSet: (payload: {
    space_id: number;
    name: string;
    description: string;
    schema_json: string;
  }): Promise<EvalSet> =>
    request('/create_eval_set', { method: 'POST', body: payload }),

  updateEvalSet: (payload: {
    id: number;
    name: string;
    description: string;
    schema_json: string;
  }): Promise<EvalSet> =>
    request('/update_eval_set', { method: 'POST', body: payload }),

  deleteEvalSet: (id: number): Promise<SuccessResp> =>
    request('/delete_eval_set', { method: 'POST', body: { id } }),

  getEvalSet: (id: number): Promise<EvalSet> =>
    request('/get_eval_set', { params: { id } }),

  listEvalSets: (page: number, pageSize: number): Promise<Paged<EvalSet>> =>
    request('/list_eval_sets', { params: { page, page_size: pageSize } }),

  addEvalSetItems: (payload: {
    eval_set_id: number;
    items: { data_json: string }[];
  }): Promise<EvalSetItem[]> =>
    request('/add_eval_set_items', { method: 'POST', body: payload }),

  deleteEvalSetItems: (payload: {
    eval_set_id: number;
    item_ids: number[];
  }): Promise<SuccessResp> =>
    request('/delete_eval_set_items', { method: 'POST', body: payload }),

  updateEvalSetItem: (payload: {
    id: number;
    data_json: string;
  }): Promise<SuccessResp> =>
    request('/update_eval_set_item', { method: 'POST', body: payload }),

  listEvalSetItems: (
    evalSetId: number,
    page: number,
    pageSize: number,
  ): Promise<Paged<EvalSetItem>> =>
    request('/list_eval_set_items', {
      params: { eval_set_id: evalSetId, page, page_size: pageSize },
    }),

  parseEvalSetFile: async (file: File): Promise<ParsedImportFile> => {
    const form = new FormData();
    form.append('file', file);
    const resp = await fetch(`${BASE}/parse_eval_set_file`, {
      method: 'POST',
      body: form,
      credentials: 'same-origin',
    });
    const data = (await resp.json()) as
      | ParsedImportFile
      | { code: number; msg: string };
    if (!resp.ok) {
      throw new Error((data as { msg?: string })?.msg || `HTTP ${resp.status}`);
    }
    const body = data as { code?: number; msg?: string };
    if (typeof body.code === 'number' && body.code !== 0) {
      throw new Error(body.msg || '文件解析失败');
    }
    return data as ParsedImportFile;
  },

  createEvaluator: (payload: {
    space_id: number;
    name: string;
    description: string;
    type: number;
    model_id: string;
    prompt: string;
    temperature: number;
  }): Promise<Evaluator> =>
    request('/create_evaluator', { method: 'POST', body: payload }),

  updateEvaluator: (payload: {
    id: number;
    name: string;
    description: string;
    model_id: string;
    prompt: string;
    temperature: number;
  }): Promise<Evaluator> =>
    request('/update_evaluator', { method: 'POST', body: payload }),

  deleteEvaluator: (id: number): Promise<SuccessResp> =>
    request('/delete_evaluator', { method: 'POST', body: { id } }),

  getEvaluator: (id: number): Promise<Evaluator> =>
    request('/get_evaluator', { params: { id } }),

  listEvaluators: (page: number, pageSize: number): Promise<Paged<Evaluator>> =>
    request('/list_evaluators', { params: { page, page_size: pageSize } }),

  createExperiment: (payload: {
    space_id: number;
    name: string;
    description: string;
    eval_set_id: number;
    target_type: number;
    target_id: string;
    target_config: string;
    evaluator_ids: number[];
    concurrency: number;
  }): Promise<Experiment> =>
    request('/create_experiment', { method: 'POST', body: payload }),

  updateExperiment: (payload: {
    id: number;
    name: string;
    description: string;
  }): Promise<Experiment> =>
    request('/update_experiment', { method: 'POST', body: payload }),

  deleteExperiment: (id: number): Promise<SuccessResp> =>
    request('/delete_experiment', { method: 'POST', body: { id } }),

  getExperiment: (id: number): Promise<Experiment> =>
    request('/get_experiment', { params: { id } }),

  listExperiments: (page: number, pageSize: number): Promise<Paged<Experiment>> =>
    request('/list_experiments', { params: { page, page_size: pageSize } }),

  startExperiment: (id: number): Promise<SuccessResp> =>
    request('/start_experiment', { method: 'POST', body: { id } }),

  getExperimentDetail: (
    id: number,
    page: number,
    pageSize: number,
  ): Promise<ExperimentDetail> =>
    request('/get_experiment_detail', {
      params: { id, page, page_size: pageSize },
    }),

  listModels: (): Promise<ModelOption[]> => listConfiguredModels(),

  listEvaluatorTemplates: (): Promise<EvaluatorTemplate[]> =>
    request<{ list: EvaluatorTemplate[] }>('/list_evaluator_templates').then(
      res => res.list,
    ),

  createEvaluatorTemplate: (payload: {
    name: string;
    description: string;
    prompt: string;
  }): Promise<{ id: number }> =>
    request('/create_evaluator_template', { method: 'POST', body: payload }),

  deleteEvaluatorTemplate: (id: number): Promise<SuccessResp> =>
    request('/delete_evaluator_template', { method: 'POST', body: { id } }),

  listTargets: (type: number): Promise<TargetInfo[]> =>
    request<{ list: TargetInfo[] }>('/list_targets', {
      params: { type },
    }).then(res => res.list),
};

async function listConfiguredModels(): Promise<ModelOption[]> {
  try {
    const resp = await fetch('/api/admin/config/model/list', {
      method: 'GET',
      credentials: 'same-origin',
    });
    if (!resp.ok) {
      return [];
    }
    const data = (await resp.json()) as {
      provider_model_list?: {
        model_list?: { id: number; display_info?: { name?: string } }[];
      }[];
    };
    const options: ModelOption[] = [];
    data.provider_model_list?.forEach(provider => {
      provider.model_list?.forEach(model => {
        if (model.id) {
          const name = model.display_info?.name?.trim();
          options.push({
            label: name ? `${name} (${model.id})` : `${model.id}`,
            value: `${model.id}`,
            name,
          });
        }
      });
    });
    return options;
  } catch {
    return [];
  }
}
