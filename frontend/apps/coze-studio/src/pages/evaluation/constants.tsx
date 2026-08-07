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

import { type TagColor } from '@coze-arch/bot-semi';

import {
  ExperimentStatus,
  ItemResultStatus,
  TargetType,
  type RunStats,
} from './types';

export const EVAL_SET_STATUS_MAP: Record<number, { label: string; color: TagColor }> = {
  [0]: { label: '草稿', color: 'grey' },
  [1]: { label: '启用', color: 'green' },
};

export const EVALUATOR_STATUS_MAP: Record<number, { label: string; color: TagColor }> = {
  [0]: { label: '草稿', color: 'grey' },
  [1]: { label: '启用', color: 'green' },
};

export const EXPERIMENT_STATUS_MAP: Record<
  number,
  { label: string; color: TagColor }
> = {
  [ExperimentStatus.PENDING]: { label: '待运行', color: 'grey' },
  [ExperimentStatus.RUNNING]: { label: '运行中', color: 'blue' },
  [ExperimentStatus.SUCCESS]: { label: '成功', color: 'green' },
  [ExperimentStatus.FAILED]: { label: '失败', color: 'red' },
  [ExperimentStatus.PARTIAL]: { label: '部分成功', color: 'orange' },
};

export const ITEM_RESULT_STATUS_MAP: Record<
  number,
  { label: string; color: TagColor }
> = {
  [ItemResultStatus.PENDING]: { label: '待执行', color: 'grey' },
  [ItemResultStatus.SUCCESS]: { label: '成功', color: 'green' },
  [ItemResultStatus.FAILED]: { label: '失败', color: 'red' },
};

export const TARGET_TYPE_MAP: Record<number, string> = {
  [TargetType.AGENT]: 'Agent',
  [TargetType.WORKFLOW]: '工作流',
  [TargetType.CHATFLOW]: '对话流',
};

export const TARGET_TYPE_OPTIONS = [
  { label: TARGET_TYPE_MAP[TargetType.AGENT], value: TargetType.AGENT },
  { label: TARGET_TYPE_MAP[TargetType.WORKFLOW], value: TargetType.WORKFLOW },
  { label: TARGET_TYPE_MAP[TargetType.CHATFLOW], value: TargetType.CHATFLOW },
];

export function parseRunStats(raw?: string): RunStats | null {
  if (!raw) {
    return null;
  }
  try {
    return JSON.parse(raw) as RunStats;
  } catch {
    return null;
  }
}

export function parseEvaluatorResults(
  raw?: string,
): { evaluator_id: number; score: number; reasoning: string }[] {
  if (!raw) {
    return [];
  }
  try {
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

export function formatDateTime(ts?: string | number): string {
  if (ts === undefined || ts === null || ts === '') {
    return '-';
  }
  // epoch milliseconds (number or numeric string)
  if (typeof ts === 'number' || /^\d{10,}$/.test(String(ts))) {
    const date = new Date(Number(ts));
    if (!isNaN(date.getTime())) {
      const pad = (n: number) => String(n).padStart(2, '0');
      return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(
        date.getDate(),
      )} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(
        date.getSeconds(),
      )}`;
    }
  }
  return String(ts).replace('T', ' ').slice(0, 19);
}

export function getTargetIdLabel(record: {
  target_type: number;
  target_id: string;
}): string {
  const typeLabel = TARGET_TYPE_MAP[record.target_type] ?? '未知';
  return `${typeLabel} / ${record.target_id}`;
}
