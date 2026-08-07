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

type ExperimentAggrResultDAO struct {
	db *gorm.DB
}

func (dao *ExperimentAggrResultDAO) Create(ctx context.Context, r *model.ExperimentAggrResult) (*model.ExperimentAggrResult, error) {
	if err := dao.db.WithContext(ctx).Create(r).Error; err != nil {
		return nil, err
	}
	return r, nil
}

func (dao *ExperimentAggrResultDAO) DeleteByExperimentID(ctx context.Context, experimentID int64) error {
	return dao.db.WithContext(ctx).
		Where("experiment_id = ?", experimentID).
		Delete(&model.ExperimentAggrResult{}).Error
}

func (dao *ExperimentAggrResultDAO) GetByExperimentID(ctx context.Context, experimentID int64) (*model.ExperimentAggrResult, error) {
	var r model.ExperimentAggrResult
	err := dao.db.WithContext(ctx).Where("experiment_id = ?", experimentID).First(&r).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}
