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
	"errors"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/evaluation/internal/dal/model"
)

type ExperimentDAO struct {
	db *gorm.DB
}

func (dao *ExperimentDAO) Create(ctx context.Context, expt *model.Experiment) (*model.Experiment, error) {
	if err := dao.db.WithContext(ctx).Create(expt).Error; err != nil {
		return nil, err
	}
	return expt, nil
}

func (dao *ExperimentDAO) Update(ctx context.Context, expt *model.Experiment) error {
	return dao.db.WithContext(ctx).Model(&model.Experiment{}).Where("id = ?", expt.ID).Updates(expt).Error
}

func (dao *ExperimentDAO) UpdateStatus(ctx context.Context, id int64, status int) error {
	return dao.db.WithContext(ctx).Model(&model.Experiment{}).Where("id = ?", id).Update("status", status).Error
}

func (dao *ExperimentDAO) Delete(ctx context.Context, id int64) error {
	return dao.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Experiment{}).Error
}

func (dao *ExperimentDAO) GetByID(ctx context.Context, id int64) (*model.Experiment, error) {
	var expt model.Experiment
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&expt).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &expt, nil
}

func (dao *ExperimentDAO) List(ctx context.Context, createdBy string, page int, pageSize int) ([]*model.Experiment, int64, error) {
	var list []*model.Experiment
	var total int64
	query := dao.db.WithContext(ctx).Model(&model.Experiment{})
	if createdBy != "" {
		query = query.Where("created_by = ?", createdBy)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
