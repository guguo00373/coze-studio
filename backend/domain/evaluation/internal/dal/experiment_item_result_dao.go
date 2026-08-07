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

package dal

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/evaluation/internal/dal/model"
)

type ExperimentItemResultDAO struct {
	db *gorm.DB
}

func (dao *ExperimentItemResultDAO) Create(ctx context.Context, r *model.ExperimentItemResult) (*model.ExperimentItemResult, error) {
	if err := dao.db.WithContext(ctx).Create(r).Error; err != nil {
		return nil, err
	}
	return r, nil
}

func (dao *ExperimentItemResultDAO) Update(ctx context.Context, r *model.ExperimentItemResult) error {
	return dao.db.WithContext(ctx).Model(&model.ExperimentItemResult{}).Where("id = ?", r.ID).Updates(r).Error
}

func (dao *ExperimentItemResultDAO) DeleteByExperimentID(ctx context.Context, experimentID int64) error {
	return dao.db.WithContext(ctx).
		Where("experiment_id = ?", experimentID).
		Delete(&model.ExperimentItemResult{}).Error
}

func (dao *ExperimentItemResultDAO) List(ctx context.Context, experimentID int64, page int, pageSize int) ([]*model.ExperimentItemResult, int64, error) {
	var list []*model.ExperimentItemResult
	var total int64
	query := dao.db.WithContext(ctx).Model(&model.ExperimentItemResult{}).Where("experiment_id = ?", experimentID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (dao *ExperimentItemResultDAO) CountByStatus(ctx context.Context, experimentID int64, status int) (int64, error) {
	var total int64
	err := dao.db.WithContext(ctx).Model(&model.ExperimentItemResult{}).
		Where("experiment_id = ? AND status = ?", experimentID, status).Count(&total).Error
	return total, err
}
