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
import { ResType } from '@coze-arch/idl/plugin_develop';

import { LibraryHeader } from '../src/pages/library/components/library-header';
import { PublishStatusFilter } from '../src/pages/library/components/publish-status-filter';

vi.mock('@coze-arch/i18n', () => {
  const labels: Record<string, string> = {
    library_filter_tags_all_status: 'All',
    library_filter_tags_published: 'Published',
    Unpublished_1: 'Unpublished',
  };
  return {
    I18n: { t: (key: string) => labels[key] ?? key },
  };
});

vi.mock('@coze-arch/coze-design/icons', () => ({
  IconCozPlus: () => null,
}));

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
  Space: ({
    children,
    'data-testid': dataTestId,
  }: {
    children: ReactNode;
    'data-testid'?: string;
  }) => <div data-testid={dataTestId}>{children}</div>,
  Menu: Object.assign(
    ({ children }: { children: ReactNode }) => <div>{children}</div>,
    {
      SubMenu: ({ children }: { children: ReactNode }) => <div>{children}</div>,
    },
  ),
}));

describe('fixed resource page controls', () => {
  it('invokes the direct create callback from the right-side header button', () => {
    // Given a fixed resource page with one create action
    const onCreate = vi.fn();

    // When the resource header is rendered and its create button is clicked
    render(
      <LibraryHeader
        title="Plugin"
        entityConfigs={[
          {
            createButton: {
              label: 'Create plugin',
              icon: null,
              dataTestId: 'create-plugin',
              onClick: onCreate,
            },
            target: [ResType.Plugin],
            onItemClick: () => undefined,
            renderActions: () => null,
          },
        ]}
      />,
    );
    fireEvent.click(screen.getByTestId('create-plugin'));

    // Then the configuration-owned callback is preserved
    expect(onCreate).toHaveBeenCalledOnce();
  });

  it('renders direct All, Published, and Unpublished status controls', () => {
    // Given the published status is selected
    const onChange = vi.fn();
    render(<PublishStatusFilter value={2} onChange={onChange} />);

    // When the direct controls render
    const controls = screen.getAllByRole('button');

    // Then all statuses are visible and the selected status is pressed
    expect(controls).toHaveLength(3);
    expect(screen.getByRole('button', { name: 'All' })).toBeVisible();
    expect(screen.getByRole('button', { name: 'Published' })).toBeVisible();
    expect(screen.getByRole('button', { name: 'Unpublished' })).toBeVisible();
    expect(screen.getByTestId('workspace.library.filter.status.2')).toHaveAttribute(
      'aria-pressed',
      'true',
    );

    fireEvent.click(screen.getByTestId('workspace.library.filter.status.1'));
    expect(onChange).toHaveBeenCalledWith(1, 'Unpublished');
  });
});
