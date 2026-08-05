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

import { I18n } from '@coze-arch/i18n';
import { Button, Space } from '@coze-arch/coze-design';

import { STATUS_FILTER_OPTIONS } from '../develop-filter-options';
import { type DevelopCustomPublishStatus } from '../type';

export const DevelopPublishStatusFilter = ({
  value,
  onChange,
}: {
  value: DevelopCustomPublishStatus;
  onChange: (value: DevelopCustomPublishStatus, label: string) => void;
}) => (
  <Space spacing={4} data-testid="workspace.develop.filter.status">
    {STATUS_FILTER_OPTIONS.map(option => {
      const label = I18n.t(option.labelI18NKey);
      const selected = value === option.value;
      return (
        <Button
          key={option.value}
          size="small"
          color={selected ? 'primary' : 'secondary'}
          type={selected ? 'primary' : 'tertiary'}
          aria-pressed={selected}
          data-testid={`workspace.develop.filter.status.${option.value}`}
          onClick={() => {
            onChange(option.value, label);
          }}
        >
          {label}
        </Button>
      );
    })}
  </Space>
);
