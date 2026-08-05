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

import { describe, expect, it } from 'vitest';
import {
  ResType,
  WorkflowMode,
  type ResourceInfo,
} from '@coze-arch/idl/plugin_develop';

import {
  getDevelopType,
  getLibraryTypeFilter,
  matchesLibraryResource,
  workspacePageKinds,
} from '../src/pages/resource-page-config';
import { DevelopCustomTypeStatus } from '../src/pages/develop/type';

describe('workspace resource page configuration', () => {
  it('maps every requested destination when building the workspace navigation', () => {
    // Given the approved fixed workspace destinations
    // When the shared page-kind list is consumed
    // Then all destinations are present in navigation order
    expect(workspacePageKinds).toEqual([
      'agent',
      'app',
      'plugin',
      'workflow',
      'chatflow',
      'knowledge',
      'prompt',
      'database',
    ]);
  });

  it('uses a fixed intelligence type when opening agent and app pages', () => {
    // Given the two intelligence list destinations
    // When their fixed request type is resolved
    // Then agent and app cannot share the all-types list
    expect(getDevelopType('agent')).toBe(DevelopCustomTypeStatus.Agent);
    expect(getDevelopType('app')).toBe(DevelopCustomTypeStatus.Project);
  });

  it.each([
    ['plugin', ResType.Plugin],
    ['workflow', ResType.Workflow],
    ['chatflow', ResType.Workflow],
    ['prompt', ResType.Prompt],
    ['database', ResType.Database],
  ] as const)('uses the existing %s resource filter', (kind, resourceType) => {
    // Given a fixed resource destination
    // When its request filter is resolved
    // Then it uses the corresponding existing backend resource type
    expect(getLibraryTypeFilter(kind)[0]).toBe(resourceType);
  });

  it('keeps the knowledge subtype wildcard under one knowledge resource type', () => {
    // Given the knowledge page supports text, table, and image subtypes
    // When its fixed request filter is resolved
    // Then the first value fixes ResType and the second selects all knowledge subtypes
    expect(getLibraryTypeFilter('knowledge')).toEqual([ResType.Knowledge, -1]);
  });

  it('distinguishes chatflow from workflow by workflow subtype', () => {
    // Given workflow resources with distinct existing subtypes
    const workflow: ResourceInfo = {
      res_type: ResType.Workflow,
      res_sub_type: WorkflowMode.Workflow,
    };
    const chatflow: ResourceInfo = {
      res_type: ResType.Workflow,
      res_sub_type: WorkflowMode.ChatFlow,
    };

    // When route-specific resource predicates run
    // Then chatflow never leaks into workflow and workflow never leaks into chatflow
    expect(matchesLibraryResource('workflow', workflow)).toBe(true);
    expect(matchesLibraryResource('workflow', chatflow)).toBe(false);
    expect(matchesLibraryResource('chatflow', workflow)).toBe(false);
    expect(matchesLibraryResource('chatflow', chatflow)).toBe(true);
  });

  it('keeps imageflow on the existing workflow destination', () => {
    // Given an imageflow returned by the existing merged workflow request
    const imageflow: ResourceInfo = { res_type: ResType.Imageflow };

    // When the workflow route filters its response
    // Then the established workflow/imageflow merge remains intact
    expect(matchesLibraryResource('workflow', imageflow)).toBe(true);
  });
});
