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

import { type FC, useRef } from 'react';

import {
  BaseLibraryPage,
  useDatabaseConfig,
  usePluginConfig,
  useWorkflowConfig,
  usePromptConfig,
  useKnowledgeConfig,
} from '@coze-studio/workspace-base/library';
import { type LibraryPageKind } from '@coze-studio/workspace-base';
import { WorkflowMode } from '@coze-arch/idl/plugin_develop';
import { I18n } from '@coze-arch/i18n';

export const LibraryPage: FC<{
  spaceId: string;
  pageKind: LibraryPageKind;
}> = ({ spaceId, pageKind }) => {
  const basePageRef = useRef<{ reloadList: () => void }>(null);
  const configCommonParams = {
    spaceId,
    reloadList: () => {
      basePageRef.current?.reloadList();
    },
  };
  const { config: pluginConfig, modals: pluginModals } =
    usePluginConfig(configCommonParams);
  const { config: workflowConfig, modals: workflowModals } =
    useWorkflowConfig({
      ...configCommonParams,
      workflowMode:
        pageKind === 'chatflow'
          ? WorkflowMode.ChatFlow
          : WorkflowMode.Workflow,
    });
  const { config: knowledgeConfig, modals: knowledgeModals } =
    useKnowledgeConfig(configCommonParams);
  const { config: promptConfig, modals: promptModals } =
    usePromptConfig(configCommonParams);
  const { config: databaseConfig, modals: databaseModals } =
    useDatabaseConfig(configCommonParams);

  const pageConfig = {
    plugin: {
      config: pluginConfig,
      title: I18n.t('library_resource_type_plugin'),
    },
    workflow: {
      config: workflowConfig,
      title: I18n.t('library_resource_type_workflow'),
    },
    chatflow: {
      config: workflowConfig,
      title: I18n.t('wf_chatflow_76'),
    },
    knowledge: {
      config: knowledgeConfig,
      title: I18n.t('library_resource_type_knowledge'),
    },
    prompt: {
      config: promptConfig,
      title: I18n.t('library_resource_type_prompt'),
    },
    database: {
      config: databaseConfig,
      title: I18n.t('new_db_001'),
    },
  } satisfies Record<
    LibraryPageKind,
    { config: typeof pluginConfig; title: string }
  >;
  const { config, title } = pageConfig[pageKind];

  return (
    <>
      <BaseLibraryPage
        spaceId={spaceId}
        pageKind={pageKind}
        pageTitle={title}
        ref={basePageRef}
        entityConfigs={[config]}
      />
      {pluginModals}
      {workflowModals}
      {promptModals}
      {databaseModals}
      {knowledgeModals}
    </>
  );
};
