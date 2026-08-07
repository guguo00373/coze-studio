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

export enum EvalSetStatus {
  DRAFT = 0,
  ENABLED = 1,
}

export enum EvaluatorStatus {
  DRAFT = 0,
  ENABLED = 1,
}

export enum ExperimentStatus {
  PENDING = 0,
  RUNNING = 1,
  SUCCESS = 2,
  FAILED = 3,
  PARTIAL = 4,
}

export enum ItemResultStatus {
  PENDING = 0,
  SUCCESS = 1,
  FAILED = 2,
}

export enum TargetType {
  AGENT = 1,
  WORKFLOW = 2,
  CHATFLOW = 3,
}

export interface EvalSet {
  id: number;
  space_id: number;
  name: string;
  description: string;
  schema_json: string;
  item_count: number;
  status: number;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface EvalSetItem {
  id: number;
  eval_set_id: number;
  data_json: string;
  created_at: string;
  updated_at?: string;
}

export interface Evaluator {
  id: number;
  space_id: number;
  name: string;
  description: string;
  type: number;
  model_id: string;
  prompt: string;
  temperature: number;
  status: number;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface Experiment {
  id: number;
  space_id: number;
  name: string;
  description: string;
  eval_set_id: number;
  target_type: number;
  target_id: string;
  target_config_json: string;
  evaluator_ids: string;
  concurrency: number;
  status: number;
  run_stats_json: string;
  error_msg: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface ExperimentItemResult {
  id: number;
  experiment_id: number;
  eval_set_item_id: number;
  status: number;
  input_json: string;
  actual_output: string;
  output_json: string;
  target_error_msg: string;
  evaluator_results_json: string;
  tokens_used_json: string;
  latency_ms: number;
}

export interface ExperimentDetail {
  experiment: Experiment;
  results: ExperimentItemResult[];
  total: number;
}

export interface Paged<T> {
  list: T[];
  total: number;
}

export interface ModelOption {
  label: string;
  value: string;
  name?: string;
}

export interface EvaluatorTemplate {
  key: string;
  name_zh: string;
  name_en: string;
  prompt_zh: string;
  prompt_en: string;
  is_builtin?: boolean;
}

export interface ParsedImportFile {
  headers: string[];
  rows: string[][];
  total: number;
}

export interface TargetInfo {
  target_id: string;
  name: string;
  type: number;
}

export interface RunStats {
  success_count?: number;
  failed_count?: number;
  scores?: Record<string, { average: number; count: number }>;
}

export interface EvaluatorResult {
  evaluator_id: number;
  score: number;
  reasoning: string;
}
