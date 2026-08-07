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

import { NavLink, Outlet } from 'react-router-dom';

const NAV_ITEMS = [
  { to: '/evaluation/sets', label: '评测集' },
  { to: '/evaluation/evaluators', label: '评估器' },
  { to: '/evaluation/experiments', label: '实验' },
];

const navLinkClass = ({ isActive }: { isActive: boolean }): string =>
  [
    'px-[16px] py-[6px] rounded-[8px] text-[14px] no-underline transition-colors',
    isActive ? 'coz-bg-primary coz-fg-plus' : 'coz-fg-secondary hover:coz-bg-primary-hovered',
  ].join(' ');

const EvaluationLayout = () => (
  <div className="h-full flex flex-col p-[24px]">
    <div className="flex items-center justify-between mb-[16px]">
      <div className="font-[500] text-[20px]">评测中心</div>
      <nav className="flex items-center gap-[8px]">
        {NAV_ITEMS.map(item => (
          <NavLink key={item.to} to={item.to} className={navLinkClass}>
            {item.label}
          </NavLink>
        ))}
      </nav>
    </div>
    <div className="flex-1 overflow-auto">
      <Outlet />
    </div>
  </div>
);

export default EvaluationLayout;
