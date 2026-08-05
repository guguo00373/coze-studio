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

import { Button, Space } from '@coze-arch/coze-design';

import { getStatusOptions, type QueryParams } from '../consts';

export const PublishStatusFilter = ({
  value,
  onChange,
}: {
  value: QueryParams['publish_status_filter'];
  onChange: (value: number, label: string) => void;
}) => (
  <Space spacing={4} data-testid="workspace.library.filter.status">
    {getStatusOptions().map(option => {
      const selected = value === option.value;
      return (
        <Button
          key={option.value}
          size="small"
          color={selected ? 'primary' : 'secondary'}
          type={selected ? 'primary' : 'tertiary'}
          aria-pressed={selected}
          data-testid={`workspace.library.filter.status.${option.value}`}
          onClick={() => {
            onChange(option.value, option.label);
          }}
        >
          {option.label}
        </Button>
      );
    })}
  </Space>
);
