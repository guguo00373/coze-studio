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

import { type ReactNode } from 'react';

import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { DevelopPublishStatusFilter } from '../src/pages/develop/components/publish-status-filter';
import {
  getDevelopFilterCacheKey,
  STATUS_FILTER_OPTIONS,
} from '../src/pages/develop/develop-filter-options';
import { getPublishRequestParam } from '../src/pages/develop/page-utils/parameters';
import {
  DevelopCustomPublishStatus,
  DevelopCustomTypeStatus,
} from '../src/pages/develop/type';

vi.mock('@coze-arch/i18n', () => {
  const labels: Record<string, string> = {
    filter_all: 'All',
    Published_1: 'Published',
    Unpublished_1: 'Unpublished',
  };
  return {
    I18n: { t: (key: string) => labels[key] ?? key },
  };
});

vi.mock('@coze-arch/coze-design', () => ({
  Button: ({
    children,
    onClick,
    'aria-pressed': ariaPressed,
    'data-testid': dataTestId,
  }: {
    children: ReactNode;
    onClick?: () => void;
    'aria-pressed'?: boolean;
    'data-testid'?: string;
  }) => (
    <button
      type="button"
      aria-pressed={ariaPressed}
      data-testid={dataTestId}
      onClick={onClick}
    >
      {children}
    </button>
  ),
  Space: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

describe('fixed Agent and App publish status', () => {
  it('maps All, Published, and Unpublished to the API contract', () => {
    // Given each direct publish status
    // When request parameters are created
    // Then All is omitted while Published and Unpublished map to booleans
    expect(getPublishRequestParam(DevelopCustomPublishStatus.All)).toBeUndefined();
    expect(getPublishRequestParam(DevelopCustomPublishStatus.Publish)).toBe(
      true,
    );
    expect(getPublishRequestParam(DevelopCustomPublishStatus.NoPublish)).toBe(
      false,
    );
  });

  it('defines exactly three status choices without recently opened', () => {
    expect(STATUS_FILTER_OPTIONS.map(option => option.value)).toEqual([
      DevelopCustomPublishStatus.All,
      DevelopCustomPublishStatus.Publish,
      DevelopCustomPublishStatus.NoPublish,
    ]);
  });

  it('isolates cached status and creator state between Agent and App', () => {
    expect(getDevelopFilterCacheKey(DevelopCustomTypeStatus.Agent)).not.toBe(
      getDevelopFilterCacheKey(DevelopCustomTypeStatus.Project),
    );
  });

  it('renders three direct accessible controls and selects Unpublished', () => {
    // Given Published is currently selected
    const onChange = vi.fn();
    render(
      <DevelopPublishStatusFilter
        value={DevelopCustomPublishStatus.Publish}
        onChange={onChange}
      />,
    );

    // When Unpublished is pressed
    const controls = screen.getAllByRole('button');
    fireEvent.click(
      screen.getByTestId(
        `workspace.develop.filter.status.${DevelopCustomPublishStatus.NoPublish}`,
      ),
    );

    // Then exactly three direct controls exist and the typed status is emitted
    expect(controls).toHaveLength(3);
    expect(
      screen.getByTestId(
        `workspace.develop.filter.status.${DevelopCustomPublishStatus.Publish}`,
      ),
    ).toHaveAttribute('aria-pressed', 'true');
    expect(onChange).toHaveBeenCalledWith(
      DevelopCustomPublishStatus.NoPublish,
      'Unpublished',
    );
  });
});
