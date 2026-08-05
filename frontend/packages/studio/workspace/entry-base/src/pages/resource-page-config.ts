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
  ResType,
  WorkflowMode,
  type ResourceInfo,
} from '@coze-arch/idl/plugin_develop';

import { DevelopCustomTypeStatus } from './develop/type';

export const workspacePageKinds = [
  'agent',
  'app',
  'plugin',
  'workflow',
  'chatflow',
  'knowledge',
  'prompt',
  'database',
] as const;

export type WorkspacePageKind = (typeof workspacePageKinds)[number];
export type DevelopPageKind = Extract<WorkspacePageKind, 'agent' | 'app'>;
export type LibraryPageKind = Exclude<WorkspacePageKind, DevelopPageKind>;

export const getDevelopType = (
  kind: DevelopPageKind,
): DevelopCustomTypeStatus =>
  kind === 'agent'
    ? DevelopCustomTypeStatus.Agent
    : DevelopCustomTypeStatus.Project;

const libraryTypeByPage = {
  plugin: ResType.Plugin,
  workflow: ResType.Workflow,
  chatflow: ResType.Workflow,
  knowledge: ResType.Knowledge,
  prompt: ResType.Prompt,
  database: ResType.Database,
} as const satisfies Record<LibraryPageKind, ResType>;

export const getLibraryTypeFilter = (kind: LibraryPageKind): number[] =>
  kind === 'knowledge'
    ? [libraryTypeByPage[kind], -1]
    : [libraryTypeByPage[kind]];

export const matchesLibraryResource = (
  kind: LibraryPageKind,
  resource: ResourceInfo,
): boolean => {
  if (kind === 'workflow') {
    return (
      resource.res_type === ResType.Imageflow ||
      (resource.res_type === ResType.Workflow &&
        resource.res_sub_type !== WorkflowMode.ChatFlow)
    );
  }
  if (kind === 'chatflow') {
    return (
      resource.res_type === ResType.Workflow &&
      resource.res_sub_type === WorkflowMode.ChatFlow
    );
  }
  return resource.res_type === libraryTypeByPage[kind];
};
